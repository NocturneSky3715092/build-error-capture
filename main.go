package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var failure BuildFailure
	flag.StringVar(&failure.Repository, "repo", "", "repository name")
	flag.StringVar(&failure.ReleaseID, "release", "", "release operation ID")
	flag.StringVar(&failure.Commit, "commit", "", "source commit")
	flag.StringVar(&failure.Stage, "stage", "", "build stage")
	flag.StringVar(&failure.DiagnosticCode, "code", "", "stable diagnostic code")
	flag.StringVar(&failure.Message, "message", "", "developer-facing diagnostic")
	flag.Parse()

	if failure.Repository == "" || failure.ReleaseID == "" || failure.Stage == "" || failure.DiagnosticCode == "" || failure.Message == "" {
		log.Fatal("repo, release, stage, code, and message are required")
	}
	apiKey := os.Getenv("INFRAI_API_KEY")
	if apiKey == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	payload, idempotencyKey := NewCapture(failure)
	data, err := (CaptureClient{APIKey: apiKey}).Capture(context.Background(), payload, idempotencyKey)
	if err != nil {
		log.Fatal(err)
	}
	var output any
	if err := json.Unmarshal(data, &output); err != nil {
		log.Fatal(err)
	}
	formatted, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(formatted))
}
