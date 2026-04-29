package inventory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type InventoryItemRequest struct {
	Condition           string                 `json:"condition,omitempty"`
	ConditionDescription string                `json:"conditionDescription,omitempty"`
	Product             *Product               `json:"product,omitempty"`
	PackageWeightAndSize *PackageWeightAndSize `json:"packageWeightAndSize,omitempty"`
	Availability        *Availability          `json:"availability,omitempty"`
	Locale              string                 `json:"locale,omitempty"`
}

type Product struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Aspects     map[string][]string `json:"aspects,omitempty"`
	Brand       string         `json:"brand,omitempty"`
	MPN         string         `json:"mpn,omitempty"`
	UPC         []string       `json:"upc,omitempty"`
	EPID        string         `json:"epid,omitempty"`
	ImageURLs   []string       `json:"imageUrls,omitempty"`
	VideoIDs    []string       `json:"videoIds,omitempty"`
}

type PackageWeightAndSize struct {
	Dimensions  *Dimensions  `json:"dimensions,omitempty"`
	PackageType string       `json:"packageType,omitempty"`
	Weight      *Weight      `json:"weight,omitempty"`
}

type Dimensions struct {
	Height float64 `json:"height,omitempty"`
	Length float64 `json:"length,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	Width  float64 `json:"width,omitempty"`
}

type Weight struct {
	Unit  string  `json:"unit,omitempty"`
	Value float64 `json:"value,omitempty"`
}

type Availability struct {
	ShipToLocationAvailability *ShipToLocationAvailability `json:"shipToLocationAvailability,omitempty"`
}

type ShipToLocationAvailability struct {
	Quantity int `json:"quantity"`
}

func (c *Client) CreateOrReplaceInventoryItem(ctx context.Context, sku string, item InventoryItemRequest) error {
	if sku == "" {
		return fmt.Errorf("inventory: sku is required")
	}
	body, _ := json.Marshal(item)
	res, err := c.doer.Do(ctx, http.MethodPut, "/inventory_item/"+url.PathEscape(sku), bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("inventory: createOrReplaceInventoryItem %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (c *Client) DeleteInventoryItem(ctx context.Context, sku string) error {
	if sku == "" {
		return fmt.Errorf("inventory: sku is required")
	}
	res, err := c.doer.Do(ctx, http.MethodDelete, "/inventory_item/"+url.PathEscape(sku), nil, "")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory: deleteInventoryItem %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

type OfferRequest struct {
	SKU                       string            `json:"sku"`
	MarketplaceID             string            `json:"marketplaceId"`
	Format                    string            `json:"format"`
	AvailableQuantity         *int              `json:"availableQuantity,omitempty"`
	CategoryID                string            `json:"categoryId,omitempty"`
	ListingDescription        string            `json:"listingDescription,omitempty"`
	ListingDuration           string            `json:"listingDuration,omitempty"`
	ListingPolicies           *ListingPolicies  `json:"listingPolicies,omitempty"`
	MerchantLocationKey       string            `json:"merchantLocationKey,omitempty"`
	PricingSummary            *PricingSummary   `json:"pricingSummary,omitempty"`
	StoreCategoryNames        []string          `json:"storeCategoryNames,omitempty"`
	Tax                       *Tax              `json:"tax,omitempty"`
	IncludeCatalogProductDetails *bool          `json:"includeCatalogProductDetails,omitempty"`
}

type ListingPolicies struct {
	PaymentPolicyID     string `json:"paymentPolicyId,omitempty"`
	ReturnPolicyID      string `json:"returnPolicyId,omitempty"`
	FulfillmentPolicyID string `json:"fulfillmentPolicyId,omitempty"`
	BestOfferTerms      *BestOfferTerms `json:"bestOfferTerms,omitempty"`
}

type BestOfferTerms struct {
	BestOfferEnabled bool    `json:"bestOfferEnabled"`
	AutoAcceptPrice  *MoneyAmount `json:"autoAcceptPrice,omitempty"`
	AutoDeclinePrice *MoneyAmount `json:"autoDeclinePrice,omitempty"`
}

type PricingSummary struct {
	Price             *MoneyAmount `json:"price,omitempty"`
	OriginallySoldFor *MoneyAmount `json:"originallySoldForRetailPriceOn,omitempty"`
}

type Tax struct {
	ApplyTax           bool   `json:"applyTax,omitempty"`
	ThirdPartyTaxCategory string `json:"thirdPartyTaxCategory,omitempty"`
	VatPercentage      float64 `json:"vatPercentage,omitempty"`
}

type CreateOfferResponse struct {
	OfferID string `json:"offerId"`
}

func (c *Client) CreateOffer(ctx context.Context, req OfferRequest) (string, error) {
	if req.SKU == "" {
		return "", fmt.Errorf("inventory: sku is required")
	}
	if req.MarketplaceID == "" {
		return "", fmt.Errorf("inventory: marketplaceId is required")
	}
	if req.Format == "" {
		req.Format = "FIXED_PRICE"
	}
	body, _ := json.Marshal(req)
	res, err := c.doer.Do(ctx, http.MethodPost, "/offer", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("inventory: createOffer %d: %s", res.StatusCode, string(res.Body))
	}
	if loc := res.Header.Get("Location"); loc != "" {
		if u, err := url.Parse(loc); err == nil {
			id := path.Base(u.Path)
			if id != "" && id != "/" && id != "." {
				return id, nil
			}
		}
	}
	var out CreateOfferResponse
	if err := json.Unmarshal(res.Body, &out); err != nil || out.OfferID == "" {
		return "", fmt.Errorf("inventory: createOffer: could not extract offerId (status %d, body %s)", res.StatusCode, string(res.Body))
	}
	return out.OfferID, nil
}

func (c *Client) UpdateOffer(ctx context.Context, offerID string, req OfferRequest) error {
	if offerID == "" {
		return fmt.Errorf("inventory: offerID is required")
	}
	body, _ := json.Marshal(req)
	res, err := c.doer.Do(ctx, http.MethodPut, "/offer/"+url.PathEscape(offerID), bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("inventory: updateOffer %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

type PublishOfferResponse struct {
	ListingID string `json:"listingId"`
}

func (c *Client) PublishOffer(ctx context.Context, offerID string) (string, error) {
	if offerID == "" {
		return "", fmt.Errorf("inventory: offerID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodPost, "/offer/"+url.PathEscape(offerID)+"/publish", nil, "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("inventory: publishOffer %d: %s", res.StatusCode, string(res.Body))
	}
	var out PublishOfferResponse
	if err := json.Unmarshal(res.Body, &out); err != nil || out.ListingID == "" {
		return "", fmt.Errorf("inventory: publishOffer: could not extract listingId (body: %s)", string(res.Body))
	}
	return out.ListingID, nil
}

func (c *Client) WithdrawOffer(ctx context.Context, offerID string) error {
	if offerID == "" {
		return fmt.Errorf("inventory: offerID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodPost, "/offer/"+url.PathEscape(offerID)+"/withdraw", nil, "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("inventory: withdrawOffer %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

func (c *Client) DeleteOffer(ctx context.Context, offerID string) error {
	if offerID == "" {
		return fmt.Errorf("inventory: offerID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodDelete, "/offer/"+url.PathEscape(offerID), nil, "")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return fmt.Errorf("inventory: deleteOffer %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

type offersList struct {
	Offers []Offer `json:"offers"`
	Total  int     `json:"total"`
	Next   string  `json:"next"`
}

// ListOffers paginates GET /sell/inventory/v1/offer filtered by
// marketplace + format, returning every offer the seller manages in
// that marketplace. Each Offer carries listingId + sku + offerId, so a
// single call recovers the full mapping for any reconciliation flow.
//
// eBay does not expose a /listing collection endpoint; this is the
// canonical way to enumerate inventory-managed listings.
func (c *Client) ListOffers(ctx context.Context, marketplaceID, format string) ([]Offer, error) {
	if marketplaceID == "" {
		return nil, fmt.Errorf("inventory: marketplaceID is required")
	}
	if format == "" {
		format = "FIXED_PRICE"
	}
	all := make([]Offer, 0)
	u := c.doer.BaseURL + "/offer?marketplace_id=" + url.QueryEscape(marketplaceID) + "&format=" + url.QueryEscape(format) + "&limit=200"
	for {
		res, err := c.doer.DoURL(ctx, http.MethodGet, u, nil, "")
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("inventory: listOffers %d: %s", res.StatusCode, string(res.Body))
		}
		var page offersList
		if err := json.Unmarshal(res.Body, &page); err != nil {
			return nil, fmt.Errorf("inventory: decode listOffers: %w", err)
		}
		all = append(all, page.Offers...)
		if page.Next == "" {
			return all, nil
		}
		u = page.Next
	}
}

func (c *Client) GetOffersBySKU(ctx context.Context, sku string) ([]Offer, error) {
	if sku == "" {
		return nil, fmt.Errorf("inventory: sku is required")
	}
	res, err := c.doer.Do(ctx, http.MethodGet, "/offer?sku="+url.QueryEscape(sku)+"&limit=100", nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory: getOffersBySKU %d: %s", res.StatusCode, string(res.Body))
	}
	var out offersList
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("inventory: decode getOffersBySKU: %w", err)
	}
	return out.Offers, nil
}
