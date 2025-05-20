# slackverifier

Go package to verify the authenticity of requests from Slack by validating their signatures and timestamps.

## Usage

```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/andreswebs/slackverifier"
)

func main() {
    signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
    if signingSecret == "" {
        log.Fatal("missing required environment variable SLACK_SIGNING_SECRET")
    }

    webhookHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        // add logic to process the request here
        // ...

        w.WriteHeader(http.StatusOK)
        return
    })

    // Wrap your handler with the Slack verification middleware
    http.Handle("/slack/webhook", slackverifier.SlackVerificationMiddleware(
        signingSecret,
        webhookHandler,
    ))

    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
```

## References

<https://docs.slack.dev/authentication/verifying-requests-from-slack>

<https://github.com/slack-go/slack/issues/353>

## Acknowledgements

<https://github.com/coro/verifyslack>

## AI Usage

Claude 3.5 Sonnet via GitHub Copilot was used to write documentation and tests for this package.

The text and the generated tests were thoroughly reviewed and updated by the author before publishing.

## Authors

**Andre Silva** - [@andreswebs](https://github.com/andreswebs)

## License

This project is licensed under the [Unlicense](UNLICENSE.md).
