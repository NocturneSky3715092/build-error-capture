package main

import "testing"

func TestNewCaptureGroupingDecision(t *testing.T) {
	tests := []struct {
		name                string
		first               BuildFailure
		second              BuildFailure
		wantSameGroup       bool
		wantSameIdempotency bool
	}{
		{
			name:                "two releases with the same build diagnosis share a group",
			first:               BuildFailure{Repository: "compiler", ReleaseID: "rel-41", Stage: "link", DiagnosticCode: "UNDEFINED_SYMBOL", Message: "missing symbol: parseConfig"},
			second:              BuildFailure{Repository: "compiler", ReleaseID: "rel-42", Stage: "link", DiagnosticCode: "UNDEFINED_SYMBOL", Message: "missing symbol: parseConfig"},
			wantSameGroup:       true,
			wantSameIdempotency: false,
		},
		{
			name:                "a retry of one release preserves its event identity",
			first:               BuildFailure{Repository: "compiler", ReleaseID: "rel-41", Stage: "test", DiagnosticCode: "RACE", Message: "race in cache test"},
			second:              BuildFailure{Repository: "compiler", ReleaseID: "rel-41", Stage: "test", DiagnosticCode: "RACE", Message: "race in cache test"},
			wantSameGroup:       true,
			wantSameIdempotency: true,
		},
		{
			name:                "different stages stay in separate groups",
			first:               BuildFailure{Repository: "compiler", ReleaseID: "rel-41", Stage: "compile", DiagnosticCode: "OOM", Message: "worker memory exceeded"},
			second:              BuildFailure{Repository: "compiler", ReleaseID: "rel-42", Stage: "package", DiagnosticCode: "OOM", Message: "worker memory exceeded"},
			wantSameGroup:       false,
			wantSameIdempotency: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstPayload, firstKey := NewCapture(tt.first)
			secondPayload, secondKey := NewCapture(tt.second)
			if got := firstPayload.Exception.Fingerprint == secondPayload.Exception.Fingerprint; got != tt.wantSameGroup {
				t.Fatalf("same group = %v, want %v", got, tt.wantSameGroup)
			}
			if got := firstKey == secondKey; got != tt.wantSameIdempotency {
				t.Fatalf("same idempotency key = %v, want %v", got, tt.wantSameIdempotency)
			}
		})
	}
}
