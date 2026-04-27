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

type Client struct {
	tokenSource auth.TokenSource
	httpClient  *http.Client
	baseURL     string
}

type Option func(*Client)

func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.baseURL = u }
}

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

type Topic struct {
	TopicID             string          `json:"topicId"`
	Description         string          `json:"description"`
	Status              string          `json:"status"`
	Context             string          `json:"context"`
	Scope               string          `json:"scope"`
	Filterable          bool            `json:"filterable"`
	AuthorizationScopes []string        `json:"authorizationScopes"`
	SupportedPayloads   []PayloadDetail `json:"supportedPayloads"`
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

type DeliveryConfig struct {
	Endpoint          string `json:"endpoint"`
	VerificationToken string `json:"verificationToken"`
}

type Destination struct {
	DestinationID  string         `json:"destinationId"`
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	DeliveryConfig DeliveryConfig `json:"deliveryConfig"`
}

type destinationListResponse struct {
	Destinations []Destination `json:"destinations"`
}

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

type CreateDestinationRequest struct {
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	DeliveryConfig DeliveryConfig `json:"deliveryConfig"`
}

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

type Subscription struct {
	SubscriptionID string `json:"subscriptionId"`
	TopicID        string `json:"topicId"`
	Status         string `json:"status"`
}

type subscriptionListResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

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

type CreateSubscriptionRequest struct {
	TopicID       string  `json:"topicId"`
	Status        string  `json:"status"`
	DestinationID string  `json:"destinationId"`
	Payload       Payload `json:"payload"`
}

type Payload struct {
	Format           string `json:"format"`
	SchemaVersion    string `json:"schemaVersion"`
	DeliveryProtocol string `json:"deliveryProtocol"`
}

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
