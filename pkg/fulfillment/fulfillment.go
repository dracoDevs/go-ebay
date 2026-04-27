// Package fulfillment wraps the eBay Sell Fulfillment API v1.
//
// Coverage is limited to GetOrder today; add ListOrders and other endpoints
// as they're needed.
package fulfillment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

const baseURL = "https://api.ebay.com/sell/fulfillment/v1"

// Client calls the Sell Fulfillment API. Construct via NewClient.
type Client struct {
	tokenSource auth.TokenSource
	httpClient  *http.Client
	baseURL     string
	marketplace string
}

// Option customizes a Client at construction.
type Option func(*Client)

// WithHTTPClient injects a custom *http.Client. Default is 30s overall timeout.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithBaseURL overrides the API base. Default is the production endpoint.
func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.baseURL = u }
}

// WithMarketplace sets the X-EBAY-C-MARKETPLACE-ID header. Default "EBAY_US".
func WithMarketplace(m string) Option {
	return func(cl *Client) { cl.marketplace = m }
}

// NewClient returns a Fulfillment API client backed by the given TokenSource.
func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		tokenSource: src,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     baseURL,
		marketplace: "EBAY_US",
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ---------- types ----------

// Order is the Sell Fulfillment v1 getOrder response. Only the fields the
// caller in this codebase consumes are surfaced; unmapped JSON is preserved
// in Raw for ad-hoc inspection.
type Order struct {
	OrderID                      string                `json:"orderId"`
	LegacyOrderID                string                `json:"legacyOrderId"`
	CreationDate                 string                `json:"creationDate"`
	LastModifiedDate             string                `json:"lastModifiedDate"`
	OrderFulfillmentStatus       string                `json:"orderFulfillmentStatus"`
	OrderPaymentStatus           string                `json:"orderPaymentStatus"`
	SellerID                     string                `json:"sellerId"`
	Buyer                        Buyer                 `json:"buyer"`
	FulfillmentStartInstructions []StartInstruction    `json:"fulfillmentStartInstructions"`
	LineItems                    []LineItem            `json:"lineItems"`
	PaymentSummary               PaymentSummary        `json:"paymentSummary"`
	PricingSummary               PricingSummary        `json:"pricingSummary"`
	TotalMarketplaceFee          Money                 `json:"totalMarketplaceFee"`
	Raw                          json.RawMessage       `json:"-"`
}

type Buyer struct {
	Username string `json:"username"`
}

type StartInstruction struct {
	FulfillmentInstructionsType string   `json:"fulfillmentInstructionsType"`
	ShippingStep                ShipStep `json:"shippingStep"`
}

type ShipStep struct {
	ShipTo ShipTo `json:"shipTo"`
}

type ShipTo struct {
	FullName       string         `json:"fullName"`
	Email          string         `json:"email"`
	ContactAddress ContactAddress `json:"contactAddress"`
	PrimaryPhone   Phone          `json:"primaryPhone"`
}

type ContactAddress struct {
	AddressLine1    string `json:"addressLine1"`
	AddressLine2    string `json:"addressLine2"`
	City            string `json:"city"`
	StateOrProvince string `json:"stateOrProvince"`
	PostalCode      string `json:"postalCode"`
	CountryCode     string `json:"countryCode"`
}

type Phone struct {
	PhoneNumber string `json:"phoneNumber"`
}

type LineItem struct {
	LineItemID                string         `json:"lineItemId"`
	LegacyItemID              string         `json:"legacyItemId"`
	Title                     string         `json:"title"`
	Quantity                  int            `json:"quantity"`
	LineItemCost              Money          `json:"lineItemCost"`
	Total                     Money          `json:"total"`
	LineItemFulfillmentStatus string         `json:"lineItemFulfillmentStatus"`
	SoldFormat                string         `json:"soldFormat"`
	Properties                LineItemProps  `json:"properties"`
}

type LineItemProps struct {
	FromBestOffer     bool `json:"fromBestOffer"`
	BuyerProtection   bool `json:"buyerProtection"`
	SoldViaAdCampaign bool `json:"soldViaAdCampaign"`
}

type PaymentSummary struct {
	TotalDueSeller Money `json:"totalDueSeller"`
}

type PricingSummary struct {
	PriceSubtotal Money `json:"priceSubtotal"`
	Total         Money `json:"total"`
}

// Money is the {value,currency} envelope eBay uses everywhere.
type Money struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// FloatValue parses Value as a float; returns 0 on empty/parse error.
func (m Money) FloatValue() float64 {
	if m.Value == "" {
		return 0
	}
	var f float64
	_, _ = fmt.Sscanf(m.Value, "%f", &f)
	return f
}

// ---------- methods ----------

// GetOrder fetches one fulfillment order by id.
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("fulfillment: orderID is required")
	}

	tok, err := c.tokenSource.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("fulfillment: token: %w", err)
	}

	u := c.baseURL + "/order/" + url.PathEscape(orderID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("fulfillment: build getOrder request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", c.marketplace)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fulfillment: getOrder http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fulfillment: read getOrder body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fulfillment: getOrder %d: %s", resp.StatusCode, string(body))
	}

	var out Order
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("fulfillment: decode getOrder: %w", err)
	}
	out.Raw = body
	return &out, nil
}
