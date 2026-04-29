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

func TestCreateOrReplaceInventoryItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s", r.Method)
		}
		if r.URL.Path != "/inventory_item/sku-1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["condition"] != "NEW" {
			t.Errorf("condition = %v", got["condition"])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	err := c.CreateOrReplaceInventoryItem(context.Background(), "sku-1", inventory.InventoryItemRequest{
		Condition: "NEW",
		Product: &inventory.Product{
			Title: "Test", ImageURLs: []string{"https://x.example.com/1.jpg"},
		},
		Availability: &inventory.Availability{
			ShipToLocationAvailability: &inventory.ShipToLocationAvailability{Quantity: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrReplaceInventoryItem: %v", err)
	}
}

func TestCreateOfferReturnsLocationID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["sku"] != "sku-1" || got["marketplaceId"] != "EBAY_US" || got["format"] != "FIXED_PRICE" {
			t.Errorf("body = %v", got)
		}
		w.Header().Set("Location", "/sell/inventory/v1/offer/OFFER-NEW")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	id, err := c.CreateOffer(context.Background(), inventory.OfferRequest{
		SKU: "sku-1", MarketplaceID: "EBAY_US",
	})
	if err != nil {
		t.Fatalf("CreateOffer: %v", err)
	}
	if id != "OFFER-NEW" {
		t.Errorf("id = %q", id)
	}
}

func TestCreateOfferReturnsBodyOfferID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"offerId":"OFFER-BODY"}`))
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	id, err := c.CreateOffer(context.Background(), inventory.OfferRequest{SKU: "s", MarketplaceID: "EBAY_US"})
	if err != nil {
		t.Fatalf("CreateOffer: %v", err)
	}
	if id != "OFFER-BODY" {
		t.Errorf("id = %q", id)
	}
}

func TestPublishOffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/offer/OFFER-1/publish" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"listingId":"LISTING-NEW"}`))
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	id, err := c.PublishOffer(context.Background(), "OFFER-1")
	if err != nil {
		t.Fatalf("PublishOffer: %v", err)
	}
	if id != "LISTING-NEW" {
		t.Errorf("id = %q", id)
	}
}

func TestWithdrawAndDeleteOffer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	if err := c.WithdrawOffer(context.Background(), "OFFER-1"); err != nil {
		t.Fatalf("WithdrawOffer: %v", err)
	}
	if err := c.DeleteOffer(context.Background(), "OFFER-1"); err != nil {
		t.Fatalf("DeleteOffer: %v", err)
	}
	if err := c.DeleteInventoryItem(context.Background(), "sku-1"); err != nil {
		t.Fatalf("DeleteInventoryItem: %v", err)
	}
}

func TestGetOffersBySKU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("sku"); got != "sku-1" {
			t.Errorf("sku = %q", got)
		}
		_, _ = w.Write([]byte(`{"offers":[{"offerId":"O1","sku":"sku-1","status":"PUBLISHED"}]}`))
	}))
	defer server.Close()

	c := inventory.NewClient(auth.StaticToken("A"), inventory.WithBaseURL(server.URL))
	offers, err := c.GetOffersBySKU(context.Background(), "sku-1")
	if err != nil {
		t.Fatalf("GetOffersBySKU: %v", err)
	}
	if len(offers) != 1 || offers[0].OfferID != "O1" {
		t.Errorf("offers = %+v", offers)
	}
}

func TestUpdateOfferRejectsEmptyID(t *testing.T) {
	c := inventory.NewClient(auth.StaticToken("A"))
	err := c.UpdateOffer(context.Background(), "", inventory.OfferRequest{})
	if err == nil || !strings.Contains(err.Error(), "offerID") {
		t.Errorf("expected offerID-required error, got %v", err)
	}
}
