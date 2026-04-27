package fulfillment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://api.ebay.com/sell/fulfillment/v1"

const (
	MarketplaceUS = "EBAY_US"
	MarketplaceGB = "EBAY_GB"
	MarketplaceCA = "EBAY_CA"
	MarketplaceAU = "EBAY_AU"
	MarketplaceDE = "EBAY_DE"
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

func WithMarketplace(m string) Option {
	return func(cl *Client) { cl.doer.DefaultHeaders["X-EBAY-C-MARKETPLACE-ID"] = m }
}

func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		doer: rest.Doer{
			TokenSource:    src,
			HTTPClient:     &http.Client{Timeout: 30 * time.Second},
			BaseURL:        baseURL,
			ErrPrefix:      "fulfillment:",
			DefaultHeaders: map[string]string{"X-EBAY-C-MARKETPLACE-ID": MarketplaceUS},
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Order struct {
	OrderID                      string             `json:"orderId"`
	LegacyOrderID                string             `json:"legacyOrderId"`
	CreationDate                 string             `json:"creationDate"`
	LastModifiedDate             string             `json:"lastModifiedDate"`
	OrderFulfillmentStatus       string             `json:"orderFulfillmentStatus"`
	OrderPaymentStatus           string             `json:"orderPaymentStatus"`
	SellerID                     string             `json:"sellerId"`
	Buyer                        Buyer              `json:"buyer"`
	FulfillmentStartInstructions []StartInstruction `json:"fulfillmentStartInstructions"`
	LineItems                    []LineItem         `json:"lineItems"`
	PaymentSummary               PaymentSummary     `json:"paymentSummary"`
	PricingSummary               PricingSummary     `json:"pricingSummary"`
	TotalMarketplaceFee          Money              `json:"totalMarketplaceFee"`
	Raw                          json.RawMessage    `json:"-"`
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
	LineItemID                string        `json:"lineItemId"`
	LegacyItemID              string        `json:"legacyItemId"`
	Title                     string        `json:"title"`
	Quantity                  int           `json:"quantity"`
	LineItemCost              Money         `json:"lineItemCost"`
	Total                     Money         `json:"total"`
	LineItemFulfillmentStatus string        `json:"lineItemFulfillmentStatus"`
	SoldFormat                string        `json:"soldFormat"`
	Properties                LineItemProps `json:"properties"`
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

type Money struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

func (m Money) FloatValue() float64 {
	if m.Value == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(m.Value, 64)
	return f
}

func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("fulfillment: orderID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodGet, "/order/"+url.PathEscape(orderID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fulfillment: getOrder %d: %s", res.StatusCode, string(res.Body))
	}
	var out Order
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("fulfillment: decode getOrder: %w", err)
	}
	out.Raw = res.Body
	return &out, nil
}
