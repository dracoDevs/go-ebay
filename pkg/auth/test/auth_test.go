package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestStaticTokenReturnsLiteral(t *testing.T) {
	got, err := auth.StaticToken("hello").Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStaticTokenEmptyErrors(t *testing.T) {
	if _, err := auth.StaticToken("").Token(context.Background()); err == nil {
		t.Error("expected error for empty static token")
	}
}

func TestRefreshTokenSourceMintsAndCaches(t *testing.T) {
	var hits int64
	server := mockTokenServer(t, &hits, func(form map[string]string) {
		if form["grant_type"] != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", form["grant_type"])
		}
		if form["refresh_token"] != "RT" {
			t.Errorf("refresh_token = %q, want RT", form["refresh_token"])
		}
		if form["scope"] == "" {
			t.Error("scope is empty")
		}
	})
	defer server.Close()

	src := newRefreshSource(server.URL, "RT", []string{auth.ScopeBase, auth.ScopeSellInventory})

	tok1, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("first Token: %v", err)
	}
	if tok1 != "ACCESS_1" {
		t.Errorf("first token = %q", tok1)
	}

	tok2, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if tok2 != "ACCESS_1" {
		t.Errorf("second token = %q (cache miss)", tok2)
	}
	if got := atomic.LoadInt64(&hits); got != 1 {
		t.Errorf("hit count = %d, want 1", got)
	}
}

func TestRefreshTokenSourceInvalidateForcesReMint(t *testing.T) {
	var hits int64
	server := mockTokenServer(t, &hits, nil)
	defer server.Close()

	src := newRefreshSource(server.URL, "RT", auth.ScopesInventory)
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("Token: %v", err)
	}

	src.Invalidate()

	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if got := atomic.LoadInt64(&hits); got != 2 {
		t.Errorf("hit count = %d, want 2", got)
	}
}

func TestRefreshTokenSourceMissingFieldsErrors(t *testing.T) {
	cases := []struct {
		name string
		src  *auth.RefreshTokenSource
	}{
		{"no AppID", &auth.RefreshTokenSource{CertID: "c", RefreshToken: "rt", Scopes: auth.ScopesInventory}},
		{"no CertID", &auth.RefreshTokenSource{AppID: "a", RefreshToken: "rt", Scopes: auth.ScopesInventory}},
		{"no RefreshToken", &auth.RefreshTokenSource{AppID: "a", CertID: "c", Scopes: auth.ScopesInventory}},
		{"no Scopes", &auth.RefreshTokenSource{AppID: "a", CertID: "c", RefreshToken: "rt"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := tc.src.Token(context.Background()); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestClientCredentialsSourceMintsAndCaches(t *testing.T) {
	var hits int64
	server := mockTokenServer(t, &hits, func(form map[string]string) {
		if form["grant_type"] != "client_credentials" {
			t.Errorf("grant_type = %q", form["grant_type"])
		}
	})
	defer server.Close()

	src := &auth.ClientCredentialsSource{
		AppID:      "a",
		CertID:     "c",
		Scope:      auth.ScopeBase,
		HTTPClient: rewriteHostClient(server.URL),
	}

	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("first Token: %v", err)
	}
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if got := atomic.LoadInt64(&hits); got != 1 {
		t.Errorf("hit count = %d, want 1", got)
	}
}

func TestClientCredentialsSourceInvalidateForcesReMint(t *testing.T) {
	var hits int64
	server := mockTokenServer(t, &hits, nil)
	defer server.Close()

	src := &auth.ClientCredentialsSource{
		AppID: "a", CertID: "c", Scope: auth.ScopeBase,
		HTTPClient: rewriteHostClient(server.URL),
	}
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("Token: %v", err)
	}
	src.Invalidate()
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if got := atomic.LoadInt64(&hits); got != 2 {
		t.Errorf("hit count = %d", got)
	}
}

func mockTokenServer(t *testing.T, hits *int64, formCheck func(map[string]string)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity/v1/oauth2/token" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Error("missing Basic auth header")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if formCheck != nil {
			form := make(map[string]string, len(r.Form))
			for k, v := range r.Form {
				if len(v) > 0 {
					form[k] = v[0]
				}
			}
			formCheck(form)
		}
		n := atomic.AddInt64(hits, 1)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"ACCESS_%d","expires_in":7200,"token_type":"Bearer"}`, n)
	}))
}

func rewriteHostClient(serverURL string) *http.Client {
	return &http.Client{Transport: &hostRewrite{serverURL: serverURL}}
}

type hostRewrite struct{ serverURL string }

func (h *hostRewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	overrideURL := h.serverURL + req.URL.Path
	if req.URL.RawQuery != "" {
		overrideURL += "?" + req.URL.RawQuery
	}
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, overrideURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header.Clone()
	return http.DefaultTransport.RoundTrip(newReq)
}

func newRefreshSource(serverURL, refreshToken string, scopes []string) *auth.RefreshTokenSource {
	return &auth.RefreshTokenSource{
		AppID:        "a",
		CertID:       "c",
		RefreshToken: refreshToken,
		Scopes:       scopes,
		HTTPClient:   rewriteHostClient(serverURL),
	}
}
