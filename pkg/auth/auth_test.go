package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStaticTokenReturnsLiteral(t *testing.T) {
	src := StaticToken("hello")
	got, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStaticTokenEmptyErrors(t *testing.T) {
	_, err := StaticToken("").Token(context.Background())
	if err == nil {
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

	src := newRefreshSourceForTest(server.URL, "RT", []string{ScopeBase, ScopeSellInventory})

	tok1, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("first Token: %v", err)
	}
	if tok1 != "ACCESS_1" {
		t.Errorf("first token = %q, want ACCESS_1", tok1)
	}

	// Second call within expiry should hit cache, not the server.
	tok2, err := src.Token(context.Background())
	if err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if tok2 != "ACCESS_1" {
		t.Errorf("second token = %q, want cached ACCESS_1", tok2)
	}
	if got := atomic.LoadInt64(&hits); got != 1 {
		t.Errorf("hit count = %d, want 1 (second call should be cached)", got)
	}
}

func TestRefreshTokenSourceRefreshesNearExpiry(t *testing.T) {
	var hits int64
	server := mockTokenServer(t, &hits, nil)
	defer server.Close()

	src := newRefreshSourceForTest(server.URL, "RT", ScopesInventory)
	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("Token: %v", err)
	}

	// Force the cache to be near expiry; the next call must re-mint.
	src.cachedExp = time.Now().Add(expiryBuffer / 2)

	if _, err := src.Token(context.Background()); err != nil {
		t.Fatalf("second Token: %v", err)
	}
	if got := atomic.LoadInt64(&hits); got != 2 {
		t.Errorf("hit count = %d, want 2 (second call should re-mint)", got)
	}
}

func TestRefreshTokenSourceMissingFieldsErrors(t *testing.T) {
	cases := []struct {
		name string
		src  *RefreshTokenSource
	}{
		{"no AppID", &RefreshTokenSource{CertID: "c", RefreshToken: "rt", Scopes: ScopesInventory}},
		{"no CertID", &RefreshTokenSource{AppID: "a", RefreshToken: "rt", Scopes: ScopesInventory}},
		{"no RefreshToken", &RefreshTokenSource{AppID: "a", CertID: "c", Scopes: ScopesInventory}},
		{"no Scopes", &RefreshTokenSource{AppID: "a", CertID: "c", RefreshToken: "rt"}},
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
			t.Errorf("grant_type = %q, want client_credentials", form["grant_type"])
		}
	})
	defer server.Close()

	src := &ClientCredentialsSource{
		AppID:      "a",
		CertID:     "c",
		Scope:      ScopeBase,
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

// ---------- helpers ----------

// mockTokenServer returns an httptest server that serves the eBay OAuth token
// endpoint shape. Each successful exchange increments hits and returns a
// distinct access token (ACCESS_1, ACCESS_2, ...) so the test can assert
// caching behavior.
func mockTokenServer(t *testing.T, hits *int64, formCheck func(map[string]string)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/identity/v1/oauth2/token" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", got)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic auth header")
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

// rewriteHostClient returns an http.Client whose Transport rewrites all
// requests to point at the test server's host. We do this instead of
// changing the package-level tokenEndpoint constant so the production URL
// stays canonical.
func rewriteHostClient(serverURL string) *http.Client {
	return &http.Client{Transport: &hostRewrite{serverURL: serverURL}}
}

type hostRewrite struct {
	serverURL string
}

func (h *hostRewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite scheme + host while preserving path and query; the token
	// endpoint is path /identity/v1/oauth2/token.
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

func newRefreshSourceForTest(serverURL, refreshToken string, scopes []string) *RefreshTokenSource {
	return &RefreshTokenSource{
		AppID:        "a",
		CertID:       "c",
		RefreshToken: refreshToken,
		Scopes:       scopes,
		HTTPClient:   rewriteHostClient(serverURL),
	}
}
