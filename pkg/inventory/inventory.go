// Package inventory wraps the eBay Sell Inventory API v1.
//
// Coverage is limited to the endpoints we actually use today:
//   - bulk_migrate_listing       (move Trading-born listings into Inventory)
//   - bulk_update_price_quantity (revise an offer's price + quantity in place)
//   - GET /offer/{offerId}        (sanity-check an offer's state)
//   - GET /inventory_item/{sku}   (sanity-check the underlying inventory item)
//
// Add more methods as they're needed; keep the surface small.
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

// Client calls the Sell Inventory API. Construct via NewClient.
type Client struct {
	tokenSource auth.TokenSource
	httpClient  *http.Client
	baseURL     string
}

// Option customizes a Client at construction.
type Option func(*Client)

// WithHTTPClient injects a custom *http.Client. The default is an http.Client
// with a 30s timeout. Use this to share a transport with other parts of your
// app (custom dialer, proxy, etc.).
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.httpClient = c }
}

// WithBaseURL overrides the API base. Default is the production endpoint.
// Use this for sandbox or local httptest servers.
func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.baseURL = u }
}

// NewClient returns an Inventory API client backed by the given TokenSource.
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

// ---------- bulkMigrateListing ----------

// MigrateListingResult is one entry in the bulkMigrateListing response.
type MigrateListingResult struct {
	ListingID      string         `json:"listingId"`
	MarketplaceID  string         `json:"marketplaceId"`
	StatusCode     int            `json:"statusCode"`
	InventoryItems []ItemListing  `json:"inventoryItems"`
	Errors         []ErrorDetail  `json:"errors"`
	Warnings       []ErrorDetail  `json:"warnings"`
}

// ItemListing is the (sku, offerId) pair the migrate API mints for each
// successfully-migrated Trading listing.
type ItemListing struct {
	SKU     string `json:"sku"`
	OfferID string `json:"offerId"`
}

// ErrorDetail is the per-listing/per-offer error envelope returned by the
// Inventory API for both migrate and price/quantity updates.
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

// BulkMigrateListings converts up to 5 Trading-API-born listings into
// Inventory API objects (inventory item + offer). The Trading listing stays
// active; this only creates the management wrappers so the same listing can
// then be updated via BulkUpdatePriceQuantity.
//
// eBay returns 400 when every listing in the batch fails and 207 when some
// succeed and some fail. Both carry per-listing breakdowns. This method
// surfaces those breakdowns as the returned slice and only errors when the
// transport itself fails or the response can't be parsed.
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

// ---------- bulkUpdatePriceQuantity ----------

// PriceQuantityUpdate is one revise-in-place operation against an existing
// Inventory API offer. Leave Price nil to skip price; leave Quantity nil to
// skip quantity. At least one must be set.
//
// When Quantity is set, SKU must also be set so the inventory item's
// shipToLocationAvailability.quantity is bumped in the same call. Without
// this, eBay's live-listing quantity stays capped by min(offer.qty, sku.qty)
// and the revise becomes a no-op whenever the new quantity exceeds the
// current ship-to-home total.
type PriceQuantityUpdate struct {
	OfferID     string
	SKU         string
	Quantity    *int
	Price       *float64
	CurrencyISO string // defaults to "USD" when Price is set
}

// PriceQuantityResult is one entry in the bulk_update_price_quantity response.
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

// BulkUpdatePriceQuantity applies in-place price and/or quantity revisions to
// existing Inventory API offers. Returns per-offer status so callers can
// handle partial failures.
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

// ---------- GetOffer ----------

// Offer is the (intentionally partial) GET /offer/{offerId} response. eBay's
// full offer schema is huge; this carries the fields callers in this codebase
// actually inspect.
type Offer struct {
	OfferID         string                 `json:"offerId"`
	SKU             string                 `json:"sku"`
	MarketplaceID   string                 `json:"marketplaceId"`
	Format          string                 `json:"format"`
	AvailableQty    int                    `json:"availableQuantity"`
	Status          string                 `json:"status"`
	ListingID       string                 `json:"listingId"`
	Pricing         *OfferPricingSummary   `json:"pricingSummary,omitempty"`
	Raw             json.RawMessage        `json:"-"`
}

// OfferPricingSummary is the price block on an Offer.
type OfferPricingSummary struct {
	Price *MoneyAmount `json:"price,omitempty"`
}

// MoneyAmount is the {value,currency} envelope eBay returns everywhere.
type MoneyAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// GetOffer fetches one offer by id.
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

// ---------- GetInventoryItem ----------

// InventoryItem is the (intentionally partial) GET /inventory_item/{sku} response.
type InventoryItem struct {
	SKU                string                          `json:"sku"`
	Availability       *InventoryItemAvailability      `json:"availability,omitempty"`
	Condition          string                          `json:"condition,omitempty"`
	Product            *InventoryItemProduct           `json:"product,omitempty"`
	Raw                json.RawMessage                 `json:"-"`
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

// GetInventoryItem fetches one inventory item by SKU.
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

// ---------- shared transport ----------

// do is the single HTTP entrypoint for the package. It mints a token, sends
// the request, and returns the response + raw body for caller-side decoding.
// The response body is fully read and closed before returning.
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
