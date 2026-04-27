package inventory

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestBulkMigrateListingsHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bulk_migrate_listing" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ACCESS" {
			t.Errorf("auth header = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		var req bulkMigrateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		if len(req.Requests) != 2 || req.Requests[0].ListingID != "L1" || req.Requests[1].ListingID != "L2" {
			t.Errorf("got requests = %+v", req.Requests)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responses":[
			{"listingId":"L1","statusCode":200,"inventoryItems":[{"sku":"S1","offerId":"O1"}]},
			{"listingId":"L2","statusCode":200,"inventoryItems":[{"sku":"S2","offerId":"O2"}]}
		]}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	res, err := c.BulkMigrateListings(context.Background(), []string{"L1", "L2"})
	if err != nil {
		t.Fatalf("BulkMigrateListings: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d results, want 2", len(res))
	}
	if res[0].InventoryItems[0].SKU != "S1" || res[0].InventoryItems[0].OfferID != "O1" {
		t.Errorf("first = %+v", res[0])
	}
}

func TestBulkMigrateListingsBatchLimit(t *testing.T) {
	c := NewClient(auth.StaticToken("ACCESS"))
	_, err := c.BulkMigrateListings(context.Background(), []string{"a", "b", "c", "d", "e", "f"})
	if err == nil || !strings.Contains(err.Error(), "5 listings") {
		t.Errorf("expected batch-limit error, got %v", err)
	}
}

func TestBulkMigrateListingsEmpty(t *testing.T) {
	c := NewClient(auth.StaticToken("ACCESS"))
	_, err := c.BulkMigrateListings(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Errorf("expected empty-input error, got %v", err)
	}
}

func TestBulkMigrateListingsHandlesPerListingFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMultiStatus)
		_, _ = w.Write([]byte(`{"responses":[
			{"listingId":"OK","statusCode":200,"inventoryItems":[{"sku":"S","offerId":"O"}]},
			{"listingId":"FAIL","statusCode":400,"errors":[{"errorId":25001,"message":"already migrated"}]}
		]}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	res, err := c.BulkMigrateListings(context.Background(), []string{"OK", "FAIL"})
	if err != nil {
		t.Fatalf("BulkMigrateListings: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d, want 2", len(res))
	}
	if res[1].StatusCode != 400 {
		t.Errorf("second status = %d, want 400", res[1].StatusCode)
	}
	if len(res[1].Errors) != 1 || res[1].Errors[0].ErrorID != 25001 {
		t.Errorf("second errors = %+v", res[1].Errors)
	}
}

func TestBulkUpdatePriceQuantitySendsExpectedShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req bulkUpdatePQRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		if len(req.Requests) != 1 {
			t.Fatalf("want 1 request entry, got %d", len(req.Requests))
		}
		entry := req.Requests[0]
		if entry.SKU != "SKU1" {
			t.Errorf("SKU = %q", entry.SKU)
		}
		if entry.ShipToLocationAvailability == nil || entry.ShipToLocationAvailability.Quantity != 5 {
			t.Errorf("shipToLocationAvailability = %+v", entry.ShipToLocationAvailability)
		}
		if len(entry.Offers) != 1 {
			t.Fatalf("offers len = %d", len(entry.Offers))
		}
		offer := entry.Offers[0]
		if offer.OfferID != "OFFER1" {
			t.Errorf("offerId = %q", offer.OfferID)
		}
		if offer.AvailableQuantity == nil || *offer.AvailableQuantity != 5 {
			t.Errorf("availableQuantity = %+v", offer.AvailableQuantity)
		}
		if offer.Price == nil || offer.Price.Value != "9.99" || offer.Price.Currency != "USD" {
			t.Errorf("price = %+v", offer.Price)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responses":[{"statusCode":200,"offerId":"OFFER1","sku":"SKU1"}]}`))
	}))
	defer server.Close()

	qty := 5
	price := 9.99
	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	res, err := c.BulkUpdatePriceQuantity(context.Background(), []PriceQuantityUpdate{{
		OfferID:  "OFFER1",
		SKU:      "SKU1",
		Quantity: &qty,
		Price:    &price,
	}})
	if err != nil {
		t.Fatalf("BulkUpdatePriceQuantity: %v", err)
	}
	if len(res) != 1 || res[0].StatusCode != 200 {
		t.Errorf("res = %+v", res)
	}
}

func TestBulkUpdatePriceQuantityRejectsQuantityWithoutSKU(t *testing.T) {
	c := NewClient(auth.StaticToken("ACCESS"))
	qty := 1
	_, err := c.BulkUpdatePriceQuantity(context.Background(), []PriceQuantityUpdate{
		{OfferID: "O", Quantity: &qty},
	})
	if err == nil || !strings.Contains(err.Error(), "sku is required") {
		t.Errorf("expected sku-required error, got %v", err)
	}
}

func TestBulkUpdatePriceQuantityRejectsEmpty(t *testing.T) {
	c := NewClient(auth.StaticToken("ACCESS"))
	_, err := c.BulkUpdatePriceQuantity(context.Background(), []PriceQuantityUpdate{
		{OfferID: "O"},
	})
	if err == nil || !strings.Contains(err.Error(), "need quantity or price") {
		t.Errorf("expected need-q-or-p error, got %v", err)
	}
}

func TestGetOffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/offer/O1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"offerId":"O1","sku":"S1","status":"PUBLISHED","availableQuantity":3}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	offer, raw, err := c.GetOffer(context.Background(), "O1")
	if err != nil {
		t.Fatalf("GetOffer: %v", err)
	}
	if offer.OfferID != "O1" || offer.SKU != "S1" || offer.AvailableQty != 3 {
		t.Errorf("offer = %+v", offer)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty raw body")
	}
}

func TestGetInventoryItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory_item/sku-with-dashes" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sku":"sku-with-dashes","condition":"NEW"}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	it, _, err := c.GetInventoryItem(context.Background(), "sku-with-dashes")
	if err != nil {
		t.Fatalf("GetInventoryItem: %v", err)
	}
	if it.SKU != "sku-with-dashes" || it.Condition != "NEW" {
		t.Errorf("item = %+v", it)
	}
}

func TestGetOfferPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`not found`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	_, _, err := c.GetOffer(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got %v", err)
	}
}
