// Package rest is a tiny shared transport for the eBay REST API clients
// (inventory, fulfillment, notification). It centralizes token fetching,
// header setup, body read, and response packaging so each client doesn't
// re-implement the same dance.
package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dracoDevs/go-ebay/pkg/auth"
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
	tok, err := d.TokenSource.Token(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("%s token: %w", d.ErrPrefix, err)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
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
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return Result{StatusCode: resp.StatusCode, Header: resp.Header, Body: raw}, nil
}
