package test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/inventory"
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
			t.Errorf("auth = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		reqs, _ := req["requests"].([]any)
		if len(reqs) != 2 {
			t.Errorf("requests len = %d", len(reqs))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responses":[
			{"listingId":"L1","statusCode":200,"inventoryItems":[{"sku":"S1","offerId":"O1"}]},
			{"listingId":"L2","statusCode":200,"inventoryItems":[{"sku":"S2","offerId":"O2"}]}
		]}`))
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
	res, err := c.BulkMigrateListings(context.Background(), []string{"L1", "L2"})
	if err != nil {
		t.Fatalf("BulkMigrateListings: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d, want 2", len(res))
	}
	if res[0].InventoryItems[0].SKU != "S1" || res[0].InventoryItems[0].OfferID != "O1" {
		t.Errorf("first = %+v", res[0])
	}
}

func TestBulkMigrateListingsBatchLimit(t *testing.T) {
	c := inventory.NewClient(auth.StaticToken("ACCESS"))
	_, err := c.BulkMigrateListings(context.Background(), []string{"a", "b", "c", "d", "e", "f"})
	if err == nil || !strings.Contains(err.Error(), "5 listings") {
		t.Errorf("expected batch-limit error, got %v", err)
	}
}

func TestBulkMigrateListingsEmpty(t *testing.T) {
	c := inventory.NewClient(auth.StaticToken("ACCESS"))
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

	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
	res, err := c.BulkMigrateListings(context.Background(), []string{"OK", "FAIL"})
	if err != nil {
		t.Fatalf("BulkMigrateListings: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d, want 2", len(res))
	}
	if res[1].StatusCode != 400 || len(res[1].Errors) != 1 || res[1].Errors[0].ErrorID != 25001 {
		t.Errorf("second = %+v", res[1])
	}
}

func TestBulkUpdatePriceQuantitySendsExpectedShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		reqs, ok := req["requests"].([]any)
		if !ok || len(reqs) != 1 {
			t.Fatalf("requests = %v", req["requests"])
		}
		entry := reqs[0].(map[string]any)
		if entry["sku"] != "SKU1" {
			t.Errorf("sku = %v", entry["sku"])
		}
		stla, _ := entry["shipToLocationAvailability"].(map[string]any)
		if stla == nil || stla["quantity"].(float64) != 5 {
			t.Errorf("shipToLocationAvailability = %v", entry["shipToLocationAvailability"])
		}
		offers := entry["offers"].([]any)
		if len(offers) != 1 {
			t.Fatalf("offers len = %d", len(offers))
		}
		offer := offers[0].(map[string]any)
		if offer["offerId"] != "OFFER1" {
			t.Errorf("offerId = %v", offer["offerId"])
		}
		if offer["availableQuantity"].(float64) != 5 {
			t.Errorf("availableQuantity = %v", offer["availableQuantity"])
		}
		price := offer["price"].(map[string]any)
		if price["value"] != "9.99" || price["currency"] != "USD" {
			t.Errorf("price = %v", price)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responses":[{"statusCode":200,"offerId":"OFFER1","sku":"SKU1"}]}`))
	}))
	defer server.Close()

	qty := 5
	price := 9.99
	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
	res, err := c.BulkUpdatePriceQuantity(context.Background(), []inventory.PriceQuantityUpdate{{
		OfferID: "OFFER1", SKU: "SKU1", Quantity: &qty, Price: &price,
	}})
	if err != nil {
		t.Fatalf("BulkUpdatePriceQuantity: %v", err)
	}
	if len(res) != 1 || res[0].StatusCode != 200 {
		t.Errorf("res = %+v", res)
	}
}

func TestBulkUpdatePriceQuantityRejectsQuantityWithoutSKU(t *testing.T) {
	c := inventory.NewClient(auth.StaticToken("ACCESS"))
	qty := 1
	_, err := c.BulkUpdatePriceQuantity(context.Background(), []inventory.PriceQuantityUpdate{
		{OfferID: "O", Quantity: &qty},
	})
	if err == nil || !strings.Contains(err.Error(), "sku is required") {
		t.Errorf("expected sku-required error, got %v", err)
	}
}

func TestBulkUpdatePriceQuantityRejectsEmpty(t *testing.T) {
	c := inventory.NewClient(auth.StaticToken("ACCESS"))
	_, err := c.BulkUpdatePriceQuantity(context.Background(), []inventory.PriceQuantityUpdate{
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

	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
	offer, raw, err := c.GetOffer(context.Background(), "O1")
	if err != nil {
		t.Fatalf("GetOffer: %v", err)
	}
	if offer.OfferID != "O1" || offer.SKU != "S1" || offer.AvailableQty != 3 {
		t.Errorf("offer = %+v", offer)
	}
	if len(raw) == 0 {
		t.Error("expected non-empty raw")
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

	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
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

	c := inventory.NewClient(auth.StaticToken("ACCESS"), inventory.WithBaseURL(server.URL))
	_, _, err := c.GetOffer(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404, got %v", err)
	}
}
