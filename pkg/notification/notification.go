package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://api.ebay.com/commerce/notification/v1"

const (
	TopicOrderConfirmation = "ORDER_CONFIRMATION"
)

type Client struct {
	doer rest.Doer
}

type Option func(*Client)

func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.doer.HTTPClient = c }
}

func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.doer.BaseURL = u }
}

func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		doer: rest.Doer{
			TokenSource: src,
			HTTPClient:  &http.Client{Timeout: 30 * time.Second},
			BaseURL:     baseURL,
			ErrPrefix:   "notification:",
		},
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
	u := c.doer.BaseURL + "/topic?limit=100"
	for {
		res, err := c.doer.DoURL(ctx, http.MethodGet, u, nil, "")
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("notification: listTopics %d: %s", res.StatusCode, string(res.Body))
		}
		var page topicListResponse
		if err := json.Unmarshal(res.Body, &page); err != nil {
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
	res, err := c.doer.Do(ctx, http.MethodGet, "/destination?limit=100", nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notification: listDestinations %d: %s", res.StatusCode, string(res.Body))
	}
	var out destinationListResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
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
	res, err := c.doer.Do(ctx, http.MethodPost, "/destination", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("notification: createDestination %d: %s", res.StatusCode, string(res.Body))
	}
	return idFromLocation(res.Header.Get("Location"), res.Body)
}

func (c *Client) UpdateConfig(ctx context.Context, alertEmail string) error {
	body, _ := json.Marshal(map[string]string{"alertEmail": alertEmail})
	res, err := c.doer.Do(ctx, http.MethodPut, "/config", bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf("notification: updateConfig %d: %s", res.StatusCode, string(res.Body))
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
	res, err := c.doer.Do(ctx, http.MethodGet, "/subscription?limit=100", nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notification: listSubscriptions %d: %s", res.StatusCode, string(res.Body))
	}
	var out subscriptionListResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
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
	res, err := c.doer.Do(ctx, http.MethodPost, "/subscription", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("notification: createSubscription %d: %s", res.StatusCode, string(res.Body))
	}
	return idFromLocation(res.Header.Get("Location"), res.Body)
}

func (c *Client) EnableSubscription(ctx context.Context, subscriptionID string) error {
	res, err := c.doer.Do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/enable", nil, "")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification: enableSubscription %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (c *Client) DisableSubscription(ctx context.Context, subscriptionID string) error {
	res, err := c.doer.Do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/disable", nil, "")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("notification: disableSubscription %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (c *Client) TestSubscription(ctx context.Context, subscriptionID string) (string, error) {
	res, err := c.doer.Do(ctx, http.MethodPost, "/subscription/"+subscriptionID+"/test", nil, "")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("notification: testSubscription %d: %s", res.StatusCode, string(res.Body))
	}
	var out struct {
		NotificationID string `json:"notificationId"`
	}
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return "", fmt.Errorf("notification: decode testSubscription: %w (body: %s)", err, string(res.Body))
	}
	return out.NotificationID, nil
}

func idFromLocation(location string, body []byte) (string, error) {
	if location == "" {
		return "", fmt.Errorf("notification: no Location header (body: %s)", string(body))
	}
	u, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("notification: parse Location %q: %w", location, err)
	}
	id := path.Base(u.Path)
	if id == "" || id == "/" || id == "." {
		return "", fmt.Errorf("notification: Location has no id segment: %q", location)
	}
	return id, nil
}
