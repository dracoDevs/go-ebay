// Package auth abstracts how callers obtain an eBay OAuth access token.
//
// A TokenSource hides whether the token is a static literal (tests, scripts),
// a refresh-token-derived user token (web app flows), or an app-level token
// (client credentials). Library packages (inventory, fulfillment, notification,
// etc.) accept a TokenSource so they don't care where the token came from.
//
// All sources that mint tokens (RefreshTokenSource, ClientCredentialsSource)
// cache the result and only re-mint when the cached token is within
// expiryBuffer of expiring. This keeps tight loops from hammering the eBay
// identity endpoint.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenSource hands out a fresh-enough eBay OAuth access token. Callers do
// not see whether the token was cached or freshly minted. Implementations
// must be safe for concurrent use.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// expiryBuffer is the safety margin we trim off eBay's reported expires_in
// before treating a cached token as still valid. eBay tokens carry ~2h TTLs,
// so a one-minute buffer is plenty.
const expiryBuffer = 60 * time.Second

// tokenEndpoint is the production eBay OAuth token URL. Sandbox is intentionally
// not exposed yet; add a per-source override if you need it.
const tokenEndpoint = "https://api.ebay.com/identity/v1/oauth2/token"

// httpDoer matches the subset of *http.Client we use, so callers can inject
// a custom client (custom timeouts, fallback dialer, etc.). Defaults to
// http.DefaultClient.
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ---------- StaticToken ----------

// StaticToken wraps a literal access token. Useful for tests and one-off CLI
// commands that already have a token in hand. The token is returned as-is on
// every call; expiry is the caller's problem.
func StaticToken(token string) TokenSource { return staticToken(token) }

type staticToken string

func (s staticToken) Token(_ context.Context) (string, error) {
	if s == "" {
		return "", fmt.Errorf("auth: static token is empty")
	}
	return string(s), nil
}

// ---------- RefreshTokenSource ----------

// RefreshTokenSource exchanges a long-lived eBay refresh token for short-lived
// access tokens, caching the result until shortly before it expires.
type RefreshTokenSource struct {
	AppID, CertID string
	RefreshToken  string
	Scopes        []string
	HTTPClient    httpDoer // optional; defaults to http.DefaultClient

	mu         sync.Mutex
	cachedTok  string
	cachedExp  time.Time
}

// Token returns a non-expired access token, minting a new one if the cache
// is empty or close to expiry.
func (r *RefreshTokenSource) Token(ctx context.Context) (string, error) {
	if r.AppID == "" || r.CertID == "" || r.RefreshToken == "" {
		return "", fmt.Errorf("auth: AppID, CertID, RefreshToken are required")
	}
	if len(r.Scopes) == 0 {
		return "", fmt.Errorf("auth: at least one scope is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cachedTok != "" && time.Until(r.cachedExp) > expiryBuffer {
		return r.cachedTok, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", r.RefreshToken)
	form.Set("scope", strings.Join(r.Scopes, " "))

	tok, exp, err := exchangeToken(ctx, r.client(), r.AppID, r.CertID, form)
	if err != nil {
		return "", err
	}
	r.cachedTok = tok
	r.cachedExp = exp
	return tok, nil
}

func (r *RefreshTokenSource) client() httpDoer {
	if r.HTTPClient != nil {
		return r.HTTPClient
	}
	return http.DefaultClient
}

// ---------- ClientCredentialsSource ----------

// ClientCredentialsSource mints application-level tokens (no user context).
// Used for endpoints like Commerce Notification topic listing that don't
// need a user identity.
type ClientCredentialsSource struct {
	AppID, CertID string
	Scope         string
	HTTPClient    httpDoer // optional; defaults to http.DefaultClient

	mu         sync.Mutex
	cachedTok  string
	cachedExp  time.Time
}

// Token returns a non-expired application token, minting a new one if the
// cache is empty or close to expiry.
func (c *ClientCredentialsSource) Token(ctx context.Context) (string, error) {
	if c.AppID == "" || c.CertID == "" {
		return "", fmt.Errorf("auth: AppID and CertID are required")
	}
	if c.Scope == "" {
		return "", fmt.Errorf("auth: Scope is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedTok != "" && time.Until(c.cachedExp) > expiryBuffer {
		return c.cachedTok, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", c.Scope)

	tok, exp, err := exchangeToken(ctx, c.client(), c.AppID, c.CertID, form)
	if err != nil {
		return "", err
	}
	c.cachedTok = tok
	c.cachedExp = exp
	return tok, nil
}

func (c *ClientCredentialsSource) client() httpDoer {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

// ---------- shared exchange ----------

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func exchangeToken(ctx context.Context, client httpDoer, appID, certID string, form url.Values) (string, time.Time, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: build token request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(appID+":"+certID)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth: token exchange http: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("auth: token exchange %d: %s", resp.StatusCode, string(body))
	}

	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", time.Time{}, fmt.Errorf("auth: decode token response: %w", err)
	}
	if tr.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("auth: empty access_token in response: %s", string(body))
	}

	exp := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	return tr.AccessToken, exp, nil
}
