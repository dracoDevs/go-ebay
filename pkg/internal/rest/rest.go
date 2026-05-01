// Package rest is a tiny shared transport for the eBay REST API clients
// (inventory, fulfillment, notification). It centralizes token fetching,
// header setup, body read, and response packaging so each client doesn't
// re-implement the same dance.
package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

const (
	maxRetries     = 5
	baseBackoff    = 1 * time.Second
	maxBackoff     = 30 * time.Second
)

type Doer struct {
	TokenSource    auth.TokenSource
	HTTPClient     *http.Client
	BaseURL        string
	ErrPrefix      string
	DefaultHeaders map[string]string
}

type Result struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (d *Doer) Do(ctx context.Context, method, path string, body io.Reader, contentType string) (Result, error) {
	return d.DoURL(ctx, method, d.BaseURL+path, body, contentType)
}

func (d *Doer) DoURL(ctx context.Context, method, fullURL string, body io.Reader, contentType string) (Result, error) {
	bodyBytes, err := drainBody(body)
	if err != nil {
		return Result{}, fmt.Errorf("%s read body: %w", d.ErrPrefix, err)
	}

	var lastResult Result
	for attempt := 0; attempt <= maxRetries; attempt++ {
		tok, err := d.TokenSource.Token(ctx)
		if err != nil {
			return Result{}, fmt.Errorf("%s token: %w", d.ErrPrefix, err)
		}
		var reader io.Reader
		if bodyBytes != nil {
			reader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
		if err != nil {
			return Result{}, fmt.Errorf("%s build request: %w", d.ErrPrefix, err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.Header.Set("Accept", "application/json")
		for k, v := range d.DefaultHeaders {
			req.Header.Set(k, v)
		}

		resp, err := d.HTTPClient.Do(req)
		if err != nil {
			return Result{}, fmt.Errorf("%s %s %s: %w", d.ErrPrefix, method, fullURL, err)
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		lastResult = Result{StatusCode: resp.StatusCode, Header: resp.Header, Body: raw}

		if !shouldRetry(resp.StatusCode) || attempt == maxRetries {
			return lastResult, nil
		}

		wait := retryDelay(resp.Header, attempt)
		select {
		case <-ctx.Done():
			return lastResult, ctx.Err()
		case <-time.After(wait):
		}
	}
	return lastResult, nil
}

func drainBody(body io.Reader) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	return io.ReadAll(body)
}

// shouldRetry reports whether the eBay API status code is one we expect
// to be transient. 429 is rate-limit; 503 is "try again shortly" under
// eBay-side load.
func shouldRetry(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable
}

// retryDelay honors a Retry-After header when eBay sends one (in seconds);
// otherwise exponential backoff capped at maxBackoff.
func retryDelay(h http.Header, attempt int) time.Duration {
	if v := h.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			d := time.Duration(secs) * time.Second
			if d > maxBackoff {
				return maxBackoff
			}
			return d
		}
	}
	d := baseBackoff << attempt
	if d > maxBackoff {
		return maxBackoff
	}
	return d
}
