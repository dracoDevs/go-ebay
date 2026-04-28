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
	"github.com/dracoDevs/go-ebay/pkg/fulfillment"
)

func TestListOrdersFiltersAndPaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "5" {
			t.Errorf("limit = %q", got)
		}
		if got := r.URL.Query().Get("filter"); !strings.Contains(got, "creationdate:") {
			t.Errorf("filter missing creationdate: %q", got)
		}
		_, _ = w.Write([]byte(`{"orders":[{"orderId":"O1","legacyOrderId":"L1"}],"total":1}`))
	}))
	defer server.Close()

	c := fulfillment.NewClient(auth.StaticToken("A"), fulfillment.WithBaseURL(server.URL))
	out, err := c.ListOrders(context.Background(), fulfillment.OrdersFilter{
		CreationDateFrom: "2026-01-01T00:00:00Z",
		CreationDateTo:   "2026-04-01T00:00:00Z",
		Limit:            5,
	})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if len(out) != 1 || out[0].OrderID != "O1" {
		t.Errorf("orders = %+v", out)
	}
}

func TestCreateShippingFulfillment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/12-34-56/shipping_fulfillment" {
			t.Errorf("path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		_ = json.Unmarshal(body, &got)
		if got["trackingNumber"] != "TRK123" {
			t.Errorf("trackingNumber = %v", got["trackingNumber"])
		}
		w.Header().Set("Location", "/sell/fulfillment/v1/order/12-34-56/shipping_fulfillment/SF-NEW")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := fulfillment.NewClient(auth.StaticToken("A"), fulfillment.WithBaseURL(server.URL))
	id, err := c.CreateShippingFulfillment(context.Background(), "12-34-56", fulfillment.ShippingFulfillmentRequest{
		LineItems:       []fulfillment.ShippingLineItem{{LineItemID: "LI1", Quantity: 1}},
		ShippingCarrier: "USPS",
		TrackingNumber:  "TRK123",
	})
	if err != nil {
		t.Fatalf("CreateShippingFulfillment: %v", err)
	}
	if id != "SF-NEW" {
		t.Errorf("id = %q", id)
	}
}

func TestCreateShippingFulfillmentRequiresLineItems(t *testing.T) {
	c := fulfillment.NewClient(auth.StaticToken("A"))
	_, err := c.CreateShippingFulfillment(context.Background(), "O", fulfillment.ShippingFulfillmentRequest{})
	if err == nil || !strings.Contains(err.Error(), "line item") {
		t.Errorf("expected line-item error, got %v", err)
	}
}
