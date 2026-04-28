package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://apiz.ebay.com/commerce/identity/v1"

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
			ErrPrefix:   "identity:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type User struct {
	UserID            string         `json:"userId"`
	Username          string         `json:"username"`
	AccountType       string         `json:"accountType,omitempty"`
	RegistrationMarketplaceID string `json:"registrationMarketplaceId,omitempty"`
	Status            string         `json:"status,omitempty"`
	Email             string         `json:"email,omitempty"`
	IndividualAccount *IndividualAccount `json:"individualAccount,omitempty"`
	BusinessAccount   *BusinessAccount   `json:"businessAccount,omitempty"`
	Raw               json.RawMessage    `json:"-"`
}

type IndividualAccount struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

type BusinessAccount struct {
	Name    string  `json:"name,omitempty"`
	Email   string  `json:"email,omitempty"`
	Address *Address `json:"address,omitempty"`
}

type Address struct {
	AddressLine1    string `json:"addressLine1,omitempty"`
	AddressLine2    string `json:"addressLine2,omitempty"`
	City            string `json:"city,omitempty"`
	StateOrProvince string `json:"stateOrProvince,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	Country         string `json:"country,omitempty"`
}

func (c *Client) GetUser(ctx context.Context) (*User, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/user/", nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("identity: getUser %d: %s", res.StatusCode, string(res.Body))
	}
	var out User
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("identity: decode getUser: %w", err)
	}
	out.Raw = res.Body
	return &out, nil
}
