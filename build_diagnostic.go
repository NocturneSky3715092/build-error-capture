package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type BuildFailure struct {
	Repository     string `json:"repository"`
	ReleaseID      string `json:"release_id"`
	Commit         string `json:"commit"`
	Stage          string `json:"stage"`
	DiagnosticCode string `json:"diagnostic_code"`
	Message        string `json:"message"`
}

type CapturePayload struct {
	Exception BuildException `json:"exception"`
}

type BuildException struct {
	Type        string       `json:"type"`
	Value       string       `json:"value"`
	Fingerprint string       `json:"fingerprint"`
	Build       BuildFailure `json:"build"`
}

func NewCapture(failure BuildFailure) (CapturePayload, string) {
	groupInput := fmt.Sprintf("%s\x00%s\x00%s", failure.Repository, failure.Stage, failure.DiagnosticCode)
	groupHash := sha256.Sum256([]byte(groupInput))
	fingerprint := hex.EncodeToString(groupHash[:16])

	eventInput := fmt.Sprintf("%s\x00%s\x00%s", failure.Repository, failure.ReleaseID, failure.DiagnosticCode)
	eventHash := sha256.Sum256([]byte(eventInput))
	idempotencyKey := "build-error-" + hex.EncodeToString(eventHash[:])

	return CapturePayload{Exception: BuildException{
		Type:        failure.DiagnosticCode,
		Value:       failure.Message,
		Fingerprint: fingerprint,
		Build:       failure,
	}}, idempotencyKey
}
