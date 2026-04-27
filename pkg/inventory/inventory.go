package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

const baseURL = "https://api.ebay.com/sell/inventory/v1"

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

type MigrateListingResult struct {
	ListingID      string        `json:"listingId"`
	MarketplaceID  string        `json:"marketplaceId"`
	StatusCode     int           `json:"statusCode"`
	InventoryItems []ItemListing `json:"inventoryItems"`
	Errors         []ErrorDetail `json:"errors"`
	Warnings       []ErrorDetail `json:"warnings"`
}

type ItemListing struct {
	SKU     string `json:"sku"`
	OfferID string `json:"offerId"`
}

type ErrorDetail struct {
	ErrorID     int    `json:"errorId"`
	Domain      string `json:"domain"`
	Category    string `json:"category"`
	Message     string `json:"message"`
	LongMessage string `json:"longMessage"`
}

type bulkMigrateRequest struct {
	Requests []migrateListingReq `json:"requests"`
}

type migrateListingReq struct {
	ListingID string `json:"listingId"`
}

type bulkMigrateResponse struct {
	Responses []MigrateListingResult `json:"responses"`
}

// eBay returns 207 (Multi-Status) on partial success and 400 when every
// listing failed; both carry per-listing breakdowns. Surface those as the
// result slice and only error on transport failures or unparseable bodies.
func (c *Client) BulkMigrateListings(ctx context.Context, listingIDs []string) ([]MigrateListingResult, error) {
	if len(listingIDs) == 0 {
		return nil, fmt.Errorf("inventory: at least one listingId required")
	}
	if len(listingIDs) > 5 {
		return nil, fmt.Errorf("inventory: bulk_migrate_listing accepts up to 5 listings per call, got %d", len(listingIDs))
	}

	reqs := make([]migrateListingReq, 0, len(listingIDs))
	for _, id := range listingIDs {
		reqs = append(reqs, migrateListingReq{ListingID: id})
	}
	body, _ := json.Marshal(bulkMigrateRequest{Requests: reqs})

	resp, raw, err := c.do(ctx, http.MethodPost, "/bulk_migrate_listing", bytes.NewReader(body), "application/json")
	if err != nil {
		return nil, err
	}

	var out bulkMigrateResponse
	if decodeErr := json.Unmarshal(raw, &out); decodeErr == nil && len(out.Responses) > 0 {
		return out.Responses, nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("inventory: bulkMigrateListing %d: %s", resp.StatusCode, string(raw))
	}
	return out.Responses, nil
}

// PriceQuantityUpdate requires SKU when Quantity is set so the inventory
// item's shipToLocationAvailability is bumped alongside the offer; without
// that, eBay caps live qty at min(offer.qty, sku.qty) and the revise can
// silently no-op.
type PriceQuantityUpdate struct {
	OfferID     string
	SKU         string
	Quantity    *int
	Price       *float64
	CurrencyISO string
}

type PriceQuantityResult struct {
	StatusCode int           `json:"statusCode"`
	OfferID    string        `json:"offerId"`
	SKU        string        `json:"sku"`
	Errors     []ErrorDetail `json:"errors"`
	Warnings   []ErrorDetail `json:"warnings"`
}

type bulkUpdatePQRequest struct {
	Requests []priceQuantityRequestEntry `json:"requests"`
}

type priceQuantityRequestEntry struct {
	Offers                     []offerPriceQuantity        `json:"offers,omitempty"`
	SKU                        string                      `json:"sku,omitempty"`
	ShipToLocationAvailability *shipToLocationAvailability `json:"shipToLocationAvailability,omitempty"`
}

type shipToLocationAvailability struct {
	Quantity int `json:"quantity"`
}

type offerPriceQuantity struct {
	OfferID           string          `json:"offerId"`
	AvailableQuantity *int            `json:"availableQuantity,omitempty"`
	Price             *inventoryPrice `json:"price,omitempty"`
}

type inventoryPrice struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type bulkUpdatePQResponse struct {
	Responses []PriceQuantityResult `json:"responses"`
}

func (c *Client) BulkUpdatePriceQuantity(ctx context.Context, updates []PriceQuantityUpdate) ([]PriceQuantityResult, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("inventory: at least one update required")
	}

	requests := make([]priceQuantityRequestEntry, 0, len(updates))
	for _, u := range updates {
		if u.OfferID == "" {
			return nil, fmt.Errorf("inventory: offerId is required for every update")
		}
		if u.Quantity == nil && u.Price == nil {
			return nil, fmt.Errorf("inventory: offer %s: need quantity or price", u.OfferID)
		}
		if u.Quantity != nil && u.SKU == "" {
			return nil, fmt.Errorf("inventory: offer %s: sku is required when updating quantity", u.OfferID)
		}

		o := offerPriceQuantity{OfferID: u.OfferID}
		if u.Quantity != nil {
			q := *u.Quantity
			o.AvailableQuantity = &q
		}
		if u.Price != nil {
			cur := u.CurrencyISO
			if cur == "" {
				cur = "USD"
			}
			o.Price = &inventoryPrice{Value: fmt.Sprintf("%.2f", *u.Price), Currency: cur}
		}

		entry := priceQuantityRequestEntry{
			Offers: []offerPriceQuantity{o},
			SKU:    u.SKU,
		}
		if u.Quantity != nil {
			entry.ShipToLocationAvailability = &shipToLocationAvailability{Quantity: *u.Quantity}
		}
		requests = append(requests, entry)
	}

	body, _ := json.Marshal(bulkUpdatePQRequest{Requests: requests})
	resp, raw, err := c.do(ctx, http.MethodPost, "/bulk_update_price_quantity", bytes.NewReader(body), "application/json")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("inventory: bulkUpdatePriceQuantity %d: %s", resp.StatusCode, string(raw))
	}

	var out bulkUpdatePQResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("inventory: decode bulkUpdatePriceQuantity: %w (body: %s)", err, string(raw))
	}
	return out.Responses, nil
}

type Offer struct {
	OfferID       string               `json:"offerId"`
	SKU           string               `json:"sku"`
	MarketplaceID string               `json:"marketplaceId"`
	Format        string               `json:"format"`
	AvailableQty  int                  `json:"availableQuantity"`
	Status        string               `json:"status"`
	ListingID     string               `json:"listingId"`
	Pricing       *OfferPricingSummary `json:"pricingSummary,omitempty"`
	Raw           json.RawMessage      `json:"-"`
}

type OfferPricingSummary struct {
	Price *MoneyAmount `json:"price,omitempty"`
}

type MoneyAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

func (c *Client) GetOffer(ctx context.Context, offerID string) (*Offer, []byte, error) {
	if offerID == "" {
		return nil, nil, fmt.Errorf("inventory: offerID is required")
	}
	resp, raw, err := c.do(ctx, http.MethodGet, "/offer/"+url.PathEscape(offerID), nil, "")
	if err != nil {
		return nil, raw, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, raw, fmt.Errorf("inventory: getOffer %d: %s", resp.StatusCode, string(raw))
	}
	var out Offer
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, raw, fmt.Errorf("inventory: decode getOffer: %w", err)
	}
	out.Raw = raw
	return &out, raw, nil
}

type InventoryItem struct {
	SKU          string                     `json:"sku"`
	Availability *InventoryItemAvailability `json:"availability,omitempty"`
	Condition    string                     `json:"condition,omitempty"`
	Product      *InventoryItemProduct      `json:"product,omitempty"`
	Raw          json.RawMessage            `json:"-"`
}

type InventoryItemAvailability struct {
	ShipToLocationAvailability *shipToLocationAvailability `json:"shipToLocationAvailability,omitempty"`
}

type InventoryItemProduct struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Brand       string   `json:"brand,omitempty"`
	MPN         string   `json:"mpn,omitempty"`
	UPC         []string `json:"upc,omitempty"`
}

func (c *Client) GetInventoryItem(ctx context.Context, sku string) (*InventoryItem, []byte, error) {
	if sku == "" {
		return nil, nil, fmt.Errorf("inventory: sku is required")
	}
	resp, raw, err := c.do(ctx, http.MethodGet, "/inventory_item/"+url.PathEscape(sku), nil, "")
	if err != nil {
		return nil, raw, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, raw, fmt.Errorf("inventory: getInventoryItem %d: %s", resp.StatusCode, string(raw))
	}
	var out InventoryItem
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, raw, fmt.Errorf("inventory: decode getInventoryItem: %w", err)
	}
	out.Raw = raw
	return &out, raw, nil
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, contentType string) (*http.Response, []byte, error) {
	tok, err := c.tokenSource.Token(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("inventory: token: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, nil, fmt.Errorf("inventory: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("inventory: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp, raw, nil
}
