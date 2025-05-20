package main

import (
	"fmt"
	"log"

	"github.com/andreswebs/slackverifier"
)

func main() {
	version := "v0"
	// timestamp := "1577831000"
	timestamp := "1577836800"
	body := "test_body"
	signingSecret := "test_secret"

	sig, err := slackverifier.GenerateSignature(version, timestamp, []byte(body), signingSecret)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sig)
}
