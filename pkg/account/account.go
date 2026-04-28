package account

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

const baseURL = "https://api.ebay.com/sell/account/v1"

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
			ErrPrefix:   "account:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type CategoryType struct {
	Default bool   `json:"default,omitempty"`
	Name    string `json:"name"`
}

type TimeDuration struct {
	Unit  string `json:"unit"`
	Value int    `json:"value"`
}

type PaymentPolicy struct {
	PaymentPolicyID  string         `json:"paymentPolicyId,omitempty"`
	Name             string         `json:"name"`
	Description      string         `json:"description,omitempty"`
	MarketplaceID    string         `json:"marketplaceId"`
	CategoryTypes    []CategoryType `json:"categoryTypes"`
	ImmediatePay     bool           `json:"immediatePay,omitempty"`
}

type ReturnPolicy struct {
	ReturnPolicyID    string        `json:"returnPolicyId,omitempty"`
	Name              string        `json:"name"`
	Description       string        `json:"description,omitempty"`
	MarketplaceID     string        `json:"marketplaceId"`
	CategoryTypes     []CategoryType `json:"categoryTypes"`
	ReturnsAccepted   bool          `json:"returnsAccepted"`
	RefundMethod      string        `json:"refundMethod,omitempty"`
	ReturnMethod      string        `json:"returnMethod,omitempty"`
	ReturnPeriod      *TimeDuration `json:"returnPeriod,omitempty"`
	ReturnShippingCostPayer string  `json:"returnShippingCostPayer,omitempty"`
}

type FulfillmentPolicy struct {
	FulfillmentPolicyID string                  `json:"fulfillmentPolicyId,omitempty"`
	Name                string                  `json:"name"`
	Description         string                  `json:"description,omitempty"`
	MarketplaceID       string                  `json:"marketplaceId"`
	CategoryTypes       []CategoryType          `json:"categoryTypes"`
	HandlingTime        *TimeDuration           `json:"handlingTime,omitempty"`
	ShippingOptions     []ShippingOption        `json:"shippingOptions,omitempty"`
	ShipToLocations     *RegionSet              `json:"shipToLocations,omitempty"`
	GlobalShipping      bool                    `json:"globalShipping,omitempty"`
}

type ShippingOption struct {
	OptionType        string             `json:"optionType"`
	CostType          string             `json:"costType"`
	ShippingServices  []ShippingService  `json:"shippingServices,omitempty"`
}

type ShippingService struct {
	ShippingCarrierCode string  `json:"shippingCarrierCode,omitempty"`
	ShippingServiceCode string  `json:"shippingServiceCode"`
	ShippingCost        *Amount `json:"shippingCost,omitempty"`
	AdditionalShippingCost *Amount `json:"additionalShippingCost,omitempty"`
	FreeShipping        bool    `json:"freeShipping,omitempty"`
	BuyerResponsibleForShipping bool `json:"buyerResponsibleForShipping,omitempty"`
}

type RegionSet struct {
	RegionIncluded []Region `json:"regionIncluded,omitempty"`
	RegionExcluded []Region `json:"regionExcluded,omitempty"`
}

type Region struct {
	RegionName string `json:"regionName"`
	RegionType string `json:"regionType,omitempty"`
}

type paymentPolicyList struct {
	PaymentPolicies []PaymentPolicy `json:"paymentPolicies"`
}

type returnPolicyList struct {
	ReturnPolicies []ReturnPolicy `json:"returnPolicies"`
}

type fulfillmentPolicyList struct {
	FulfillmentPolicies []FulfillmentPolicy `json:"fulfillmentPolicies"`
}

func (c *Client) ListPaymentPolicies(ctx context.Context, marketplaceID string) ([]PaymentPolicy, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/payment_policy?marketplace_id="+url.QueryEscape(marketplaceID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: listPaymentPolicies %d: %s", res.StatusCode, string(res.Body))
	}
	var out paymentPolicyList
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("account: decode listPaymentPolicies: %w", err)
	}
	return out.PaymentPolicies, nil
}

func (c *Client) CreatePaymentPolicy(ctx context.Context, p PaymentPolicy) (string, error) {
	body, _ := json.Marshal(p)
	res, err := c.doer.Do(ctx, http.MethodPost, "/payment_policy", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("account: createPaymentPolicy %d: %s", res.StatusCode, string(res.Body))
	}
	return idFrom(res, "paymentPolicyId")
}

func (c *Client) ListReturnPolicies(ctx context.Context, marketplaceID string) ([]ReturnPolicy, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/return_policy?marketplace_id="+url.QueryEscape(marketplaceID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: listReturnPolicies %d: %s", res.StatusCode, string(res.Body))
	}
	var out returnPolicyList
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("account: decode listReturnPolicies: %w", err)
	}
	return out.ReturnPolicies, nil
}

func (c *Client) CreateReturnPolicy(ctx context.Context, p ReturnPolicy) (string, error) {
	body, _ := json.Marshal(p)
	res, err := c.doer.Do(ctx, http.MethodPost, "/return_policy", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("account: createReturnPolicy %d: %s", res.StatusCode, string(res.Body))
	}
	return idFrom(res, "returnPolicyId")
}

func (c *Client) ListFulfillmentPolicies(ctx context.Context, marketplaceID string) ([]FulfillmentPolicy, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/fulfillment_policy?marketplace_id="+url.QueryEscape(marketplaceID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("account: listFulfillmentPolicies %d: %s", res.StatusCode, string(res.Body))
	}
	var out fulfillmentPolicyList
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("account: decode listFulfillmentPolicies: %w", err)
	}
	return out.FulfillmentPolicies, nil
}

func (c *Client) CreateFulfillmentPolicy(ctx context.Context, p FulfillmentPolicy) (string, error) {
	body, _ := json.Marshal(p)
	res, err := c.doer.Do(ctx, http.MethodPost, "/fulfillment_policy", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("account: createFulfillmentPolicy %d: %s", res.StatusCode, string(res.Body))
	}
	return idFrom(res, "fulfillmentPolicyId")
}

func idFrom(res rest.Result, idField string) (string, error) {
	if loc := res.Header.Get("Location"); loc != "" {
		u, err := url.Parse(loc)
		if err == nil {
			id := path.Base(u.Path)
			if id != "" && id != "/" && id != "." {
				return id, nil
			}
		}
	}
	var body map[string]any
	if err := json.Unmarshal(res.Body, &body); err == nil {
		if v, ok := body[idField].(string); ok && v != "" {
			return v, nil
		}
	}
	return "", fmt.Errorf("account: could not extract %s from response (status %d, body: %s)", idField, res.StatusCode, string(res.Body))
}
