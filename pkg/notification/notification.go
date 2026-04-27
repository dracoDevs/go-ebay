// Package notification wraps the eBay Commerce Notification API v1.
//
// The API mixes app-level operations (managing destinations, listing topics)
// with user-level operations (subscribing a specific seller). Both flow
// through the same Client because they share a base URL; the difference is
// just which TokenSource you hand it (auth.ClientCredentialsSource for app
// ops, auth.RefreshTokenSource for user ops).
package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

const baseURL = "https://api.ebay.com/commerce/notification/v1"

// Client calls the Commerce Notification API. Construct via NewClient.
type Client struct {
	tokenSource auth.TokenSource
	httpClient  *http.Client
	baseURL     string
}

// Option customizes a Client at construction.
type Option func(*Client)

// WithHTTPClient injects a custom *http.Client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithBaseURL overrides the API base. Default is the production endpoint.
func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.baseURL = u }
}

// NewClient returns a Notification API client backed by the given TokenSource.
// Use a ClientCredentialsSource for app-level operations (topics, destinations,
// config). Use a RefreshTokenSource for user-level operations (subscriptions
// belong to a specific seller).
func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		tokenSource: src,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     baseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ---------- Topic ----------

// Topic is one notification topic, returned by ListTopics. App-level
// (client_credentials) auth is sufficient.
type Topic struct {
	TopicID             string           `json:"topicId"`
	Description         string           `json:"description"`
	Status              string           `json:"status"`
	Context             string           `json:"context"`
	Scope               string           `json:"scope"`
	Filterable          bool             `json:"filterable"`
	AuthorizationScopes []string         `json:"authorizationScopes"`
	SupportedPayloads   []PayloadDetail  `json:"supportedPayloads"`
}

type PayloadDetail struct {
	DeliveryProtocol string   `json:"deliveryProtocol"`
	Deprecated       bool     `json:"deprecated"`
	Format           []string `json:"format"`
	SchemaVersion    string   `json:"schemaVersion"`
}

type topicListResponse struct {
	Topics []Topic `json:"topics"`
	Next   string  `json:"next"`
}

// ListTopics fetches every topic the application is authorized to subscribe
// to, following pagination.
func (c *Client) ListTopics(ctx context.Context) ([]Topic, error) {
	all := make([]Topic, 0)
	u := c.baseURL + "/topic?limit=100"
	for {
		resp, raw, err := c.doURL(ctx, http.MethodGet, u, nil, "")
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("notification: listTopics %d: %s", resp.StatusCode, string(raw))
		}
		var page topicListResponse
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, fmt.Errorf("notification: decode topics: %w", err)
		}
		all = append(all, page.Topics...)
		if page.Next == "" {
			return all, nil
		}
		u = page.Next
	}
}

// ---------- Destination ----------

// DeliveryConfig is the where-and-how for notification delivery.
type DeliveryConfig struct {
	Endpoint          string `json:"endpoint"`
	VerificationToken string `json:"verificationToken"`
}

// Destination is a registered webhook target.
type Destination struct {
	DestinationID  string         `json:"destinationId"`
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	DeliveryConfig DeliveryConfig `json:"deliveryConfig"`
}

type destinationListResponse struct {
	Destinations []Destination `json:"destinations"`
}

// ListDestinations returns all registered destinations for the application.
// App-level auth.
func (c *Client) ListDestinations(ctx context.Context) ([]Destination, error) {
	resp, raw, err := c.do(ctx, http.MethodGet, "/destination?limit=100", nil, "")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notification: listDestinations %d: %s", resp.StatusCode, string(raw))
	}
	var out destinationListResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("notification: decode destinations: %w", err)
	}
	return out.Destinations, nil
}

// CreateDestinationRequest is the body for CreateDestination. Status is
// typically "ENABLED".
type CreateDestinationRequest struct {
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	DeliveryConfig DeliveryConfig `json:"deliveryConfig"`
}

// CreateDestination registers a new webhook target. eBay challenges the
// endpoint synchronously; if the challenge fails this call returns an error.
// Returns the new destination's id (parsed from the Location header).
func (c *Client) CreateDestination(ctx context.Context, req CreateDestinationRequest) (string, error) {
	body, _ := json.Marshal(req)
	resp, raw, err := c.do(ctx, http.MethodPost, "/destination", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("notification: createDestination %d: %s", resp.StatusCode, string(raw))
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("notification: createDestination: no Location header (body: %s)", string(raw))
	}
	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

// UpdateConfig sets account-wide notification config (currently just the
// alert email). App-level auth.
func (c *Client) UpdateConfig(ctx context.Context, alertEmail string) error {
	body, _ := json.Marshal(map[string]string{"alertEmail": alertEmail})
	resp, raw, err := c.do(ctx, http.MethodPut, "/config", bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification: updateConfig %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// ---------- Subscription ----------

// Subscription is one user-level subscription to a topic.
type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
	TopicID        string `json:"topicId"`
	Status         string `json:"status"`
}

type subscriptionListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// ListSubscriptions returns all subscriptions owned by the user whose token
// is on the request. Requires user-level auth (refresh-token flow).
func (c *Client) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	resp, raw, err := c.do(ctx, http.MethodGet, "/subscription?limit=100", nil, "")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notification: listSubscriptions %d: %s", resp.StatusCode, string(raw))
	}
	var out subscriptionListResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("notification: decode subscriptions: %w", err)
	}
	return out.Subscriptions, nil
}

// CreateSubscriptionRequest is the body for CreateSubscription. Pass status
// "ENABLED" or "DISABLED".
type CreateSubscriptionRequest struct {
	TopicID       string  `json:"topicId"`
	Status        string  `json:"status"`
	DestinationID string  `json:"destinationId"`
	Payload       Payload `json:"payload"`
}

// Payload describes the format and protocol of delivered notifications.
// Typical values: Format="JSON", SchemaVersion="1.0", DeliveryProtocol="HTTPS".
type Payload struct {
	Format           string `json:"format"`
	SchemaVersion    string `json:"schemaVersion"`
	DeliveryProtocol string `json:"deliveryProtocol"`
}

// CreateSubscription creates a new subscription for the user whose token is
// on the request. Returns the new subscription's id (parsed from Location).
func (c *Client) CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (string, error) {
	body, _ := json.Marshal(req)
	resp, raw, err := c.do(ctx, http.MethodPost, "/subscription", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("notification: createSubscription %d: %s", resp.StatusCode, string(raw))
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("notification: createSubscription: no Location header (body: %s)", string(raw))
	}
	parts := strings.Split(location, "/")
	return parts[len(parts)-1], nil
}

// EnableSubscription flips a subscription from DISABLED to ENABLED.
func (c *Client) EnableSubscription(ctx context.Context, subscriptionID string) error {
	resp, raw, err := c.do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/enable", nil, "")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification: enableSubscription %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// DisableSubscription flips a subscription from ENABLED to DISABLED.
func (c *Client) DisableSubscription(ctx context.Context, subscriptionID string) error {
	resp, raw, err := c.do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/disable", nil, "")
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification: disableSubscription %d: %s", resp.StatusCode, string(raw))
	}
	return nil
}

// TestSubscription asks eBay to fire a synthetic notification at the
// subscription's destination so you can verify your webhook handler. Returns
// the notificationId eBay assigned.
func (c *Client) TestSubscription(ctx context.Context, subscriptionID string) (string, error) {
	resp, raw, err := c.do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/test", nil, "")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("notification: testSubscription %d: %s", resp.StatusCode, string(raw))
	}
	var out struct {
		NotificationID string `json:"notificationId"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("notification: decode testSubscription: %w (body: %s)", err, string(raw))
	}
	return out.NotificationID, nil
}

// ---------- shared transport ----------

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, []byte, error) {
	return c.doURL(ctx, method, c.baseURL+path, body, contentType)
}

func (c *Client) doURL(ctx context.Context, method, fullURL string, body io.Reader, contentType string) (*http.Response, []byte, error) {
	tok, err := c.tokenSource.Token(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("notification: token: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("notification: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("notification: %s %s: %w", method, fullURL, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp, raw, nil
}
