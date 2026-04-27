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

type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

const expiryBuffer = 60 * time.Second

const tokenEndpoint = "https://api.ebay.com/identity/v1/oauth2/token"

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func StaticToken(token string) TokenSource { return staticToken(token) }

type staticToken string

func (s staticToken) Token(_ context.Context) (string, error) {
	if s == "" {
		return "", fmt.Errorf("auth: static token is empty")
	}
	return string(s), nil
}

type RefreshTokenSource struct {
	AppID, CertID string
	RefreshToken  string
	Scopes        []string
	HTTPClient    httpDoer

	mu        sync.Mutex
	cachedTok string
	cachedExp time.Time
}

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

// Invalidate forces the next Token call to re-mint. Call after a 401 from
// the API.
func (r *RefreshTokenSource) Invalidate() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cachedTok = ""
	r.cachedExp = time.Time{}
}

type ClientCredentialsSource struct {
	AppID, CertID string
	Scope         string
	HTTPClient    httpDoer

	mu        sync.Mutex
	cachedTok string
	cachedExp time.Time
}

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

func (c *ClientCredentialsSource) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedTok = ""
	c.cachedExp = time.Time{}
}

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
