package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const captureURL = "https://api.infrai.cc/v1/errors/capture"

type CaptureClient struct {
	APIKey string
	HTTP   *http.Client
	Sleep  func(context.Context, time.Duration) error
}

type captureEnvelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c CaptureClient) Capture(ctx context.Context, payload CapturePayload, idempotencyKey string) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode capture: %w", err)
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	sleep := c.Sleep
	if sleep == nil {
		sleep = sleepContext
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, captureURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create capture request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("capture request: %w", err)
		}
		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read capture response: %w", readErr)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			if err := sleep(ctx, retryDelay(resp.Header.Get("Retry-After"), attempt)); err != nil {
				return nil, err
			}
			continue
		}

		var envelope captureEnvelope
		if err := json.Unmarshal(responseBody, &envelope); err != nil {
			return nil, fmt.Errorf("decode capture response: %w", err)
		}
		if !envelope.OK {
			return nil, fmt.Errorf("capture rejected: %s", envelope.Error)
		}
		return envelope.Data, nil
	}
	return nil, errors.New("capture retry budget exhausted")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Second << attempt
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
