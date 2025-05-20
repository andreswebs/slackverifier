package slackverifier_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/andreswebs/slackverifier"
)

func TestMain(m *testing.M) {
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name          string
		data          slackverifier.SlackRequestData
		expectedOk    bool
		expectedError error
	}{
		{
			name: "Valid signature",
			data: slackverifier.SlackRequestData{
				Version:        "v0",
				RawBody:        []byte("test_body"),
				SigningSecret:  "test_secret",
				Timestamp:      "1577836800", // 2020-01-01 00:00:00 UTC
				SlackSignature: "v0=f7aa70e347182d9d30a148493fab76a0cc710481d8aadcc6291abed5f1c1d41c",
			},
			expectedOk:    true,
			expectedError: nil,
		},
		{
			name: "Invalid signature",
			data: slackverifier.SlackRequestData{
				Version:        "v0",
				RawBody:        []byte("test_body"),
				SigningSecret:  "test_secret",
				Timestamp:      "1577836800",
				SlackSignature: "v0=invalid_signature",
			},
			expectedOk:    false,
			expectedError: slackverifier.ErrInvalidSignature,
		},
		{
			name: "Empty version defaults to v0",
			data: slackverifier.SlackRequestData{
				Version:        "",
				RawBody:        []byte("test_body"),
				SigningSecret:  "test_secret",
				Timestamp:      "1577836800",
				SlackSignature: "v0=f7aa70e347182d9d30a148493fab76a0cc710481d8aadcc6291abed5f1c1d41c",
			},
			expectedOk:    true,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.data.VerifySignature()
			if ok != tt.expectedOk {
				t.Errorf("got ok=%v, want ok=%v", ok, tt.expectedOk)
			}
			if err != tt.expectedError {
				t.Errorf("got err=%v, want err=%v", err, tt.expectedError)
			}
		})
	}
}

func TestVerifyTimestamp(t *testing.T) {
	now := time.Now()
	oldTimestamp := now.Add(-2 * time.Minute).Unix()
	recentTimestamp := now.Add(-30 * time.Second).Unix()

	tests := []struct {
		name          string
		data          slackverifier.SlackRequestData
		expectedOk    bool
		expectedError error
	}{
		{
			name: "Valid recent timestamp",
			data: slackverifier.SlackRequestData{
				Timestamp:            fmt.Sprint(recentTimestamp),
				MaxAllowedRequestAge: time.Minute,
			},
			expectedOk:    true,
			expectedError: nil,
		},
		{
			name: "Expired timestamp",
			data: slackverifier.SlackRequestData{
				Timestamp:            fmt.Sprint(oldTimestamp),
				MaxAllowedRequestAge: time.Minute,
			},
			expectedOk:    false,
			expectedError: slackverifier.ErrMaxAllowedRequestAgeExceeded,
		},
		{
			name: "Default MaxAllowedRequestAge",
			data: slackverifier.SlackRequestData{
				Timestamp: fmt.Sprint(recentTimestamp),
			},
			expectedOk:    true,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.data.VerifyTimestamp()
			if ok != tt.expectedOk {
				t.Errorf("got ok=%v, want ok=%v", ok, tt.expectedOk)
			}
			if err != tt.expectedError {
				t.Errorf("got err=%v, want err=%v", err, tt.expectedError)
			}
		})
	}
}

func TestIntTimestamp(t *testing.T) {
	tests := []struct {
		name        string
		timestamp   string
		expectedTs  int64
		expectError bool
	}{
		{
			name:        "Valid timestamp current time",
			timestamp:   "1621468800", // 2021-05-20 00:00:00 UTC
			expectedTs:  1621468800,
			expectError: false,
		},
		{
			name:        "Future timestamp",
			timestamp:   "1968416400", // 2032-05-20 00:00:00 UTC
			expectedTs:  1968416400,
			expectError: false,
		},
		{
			name:        "Past timestamp",
			timestamp:   "1589932800", // 2020-05-20 00:00:00 UTC
			expectedTs:  1589932800,
			expectError: false,
		},
		{
			name:        "Invalid timestamp non-numeric",
			timestamp:   "not_a_number",
			expectedTs:  0,
			expectError: true,
		},
		{
			name:        "Empty timestamp",
			timestamp:   "",
			expectedTs:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := slackverifier.SlackRequestData{Timestamp: tt.timestamp}
			ts, err := s.IntTimestamp()
			if (err != nil) != tt.expectError {
				t.Errorf("got err=%v, expectError=%v", err, tt.expectError)
			}
			if ts != tt.expectedTs {
				t.Errorf("got ts=%v, want ts=%v", ts, tt.expectedTs)
			}
		})
	}
}

func TestSlackVerificationMiddleware(t *testing.T) {
	timestamp := fmt.Sprint(time.Now().Unix())
	secret := "test_secret"
	body := []byte("test_body")
	signature, _ := slackverifier.GenerateSignature("v0", timestamp, body, secret)

	tests := []struct {
		name          string
		method        string
		headers       map[string]string
		body          []byte
		signingSecret string
		wantStatus    int
		handlerCalled bool
	}{
		{
			name:   "Valid request",
			method: "POST",
			headers: map[string]string{
				"X-Slack-Request-Timestamp": timestamp,
				"X-Slack-Signature":         signature,
			},
			body:          body,
			signingSecret: secret,
			wantStatus:    http.StatusOK,
			handlerCalled: true,
		},
		{
			name:   "Missing signature header",
			method: "POST",
			headers: map[string]string{
				"X-Slack-Request-Timestamp": "1577836800",
			},
			body:          []byte("test_body"),
			signingSecret: "test_secret",
			wantStatus:    http.StatusBadRequest,
			handlerCalled: false,
		},
		{
			name:   "Missing timestamp header",
			method: "POST",
			headers: map[string]string{
				"X-Slack-Signature": "v0=f7aa70e347182d9d30a148493fab76a0cc710481d8aadcc6291abed5f1c1d41c",
			},
			body:          []byte("test_body"),
			signingSecret: "test_secret",
			wantStatus:    http.StatusBadRequest,
			handlerCalled: false,
		},
		{
			name:   "Invalid HTTP method",
			method: "GET",
			headers: map[string]string{
				"X-Slack-Signature":         "v0=f7aa70e347182d9d30a148493fab76a0cc710481d8aadcc6291abed5f1c1d41c",
				"X-Slack-Request-Timestamp": "1577836800",
			},
			body:          []byte("test_body"),
			signingSecret: "test_secret",
			wantStatus:    http.StatusMethodNotAllowed,
			handlerCalled: false,
		},
		{
			name:   "Invalid signature",
			method: "POST",
			headers: map[string]string{
				"X-Slack-Signature":         "v0=invalid",
				"X-Slack-Request-Timestamp": "1577836800",
			},
			body:          []byte("test_body"),
			signingSecret: "test_secret",
			wantStatus:    http.StatusUnauthorized,
			handlerCalled: false,
		},
		{
			name:   "Expired timestamp",
			method: "POST",
			headers: map[string]string{
				"X-Slack-Request-Timestamp": "1577831000",
				"X-Slack-Signature":         "v0=d965adaa4fe1bb598efad080a4f04ae4f6446e1a05d0839eab713abe9acb99bb",
			},
			body:          []byte("test_body"),
			signingSecret: "test_secret",
			wantStatus:    http.StatusUnauthorized,
			handlerCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := slackverifier.SlackVerificationMiddleware(tt.signingSecret, nextHandler)

			req := httptest.NewRequest(tt.method, "/webhook", bytes.NewReader(tt.body))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}

			if handlerCalled != tt.handlerCalled {
				t.Errorf("next handler called = %v, want %v", handlerCalled, tt.handlerCalled)
			}
		})
	}
}
