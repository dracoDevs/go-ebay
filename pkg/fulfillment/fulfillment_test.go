package fulfillment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

const sampleOrderJSON = `{
  "orderId": "12-34567-89012",
  "legacyOrderId": "L-1",
  "creationDate": "2026-04-20T10:00:00Z",
  "orderFulfillmentStatus": "NOT_STARTED",
  "orderPaymentStatus": "PAID",
  "sellerId": "seller-1",
  "buyer": {"username": "buyer-1"},
  "lineItems": [
    {"lineItemId": "LI1", "legacyItemId": "1234", "title": "Item", "quantity": 1,
     "lineItemCost": {"value": "9.99", "currency": "USD"},
     "total": {"value": "9.99", "currency": "USD"},
     "lineItemFulfillmentStatus": "NOT_STARTED"}
  ],
  "paymentSummary": {"totalDueSeller": {"value": "8.50", "currency": "USD"}},
  "pricingSummary": {"priceSubtotal": {"value": "9.99", "currency": "USD"}, "total": {"value": "9.99", "currency": "USD"}},
  "totalMarketplaceFee": {"value": "1.49", "currency": "USD"}
}`

func TestGetOrderHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/order/12-34567-89012" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ACCESS" {
			t.Errorf("auth = %q", got)
		}
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_US" {
			t.Errorf("marketplace = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleOrderJSON))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	order, err := c.GetOrder(context.Background(), "12-34567-89012")
	if err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
	if order.OrderID != "12-34567-89012" {
		t.Errorf("orderId = %q", order.OrderID)
	}
	if order.Buyer.Username != "buyer-1" {
		t.Errorf("buyer = %+v", order.Buyer)
	}
	if len(order.LineItems) != 1 || order.LineItems[0].Title != "Item" {
		t.Errorf("lineItems = %+v", order.LineItems)
	}
	if got := order.PaymentSummary.TotalDueSeller.FloatValue(); got != 8.5 {
		t.Errorf("totalDueSeller = %v, want 8.5", got)
	}
	if len(order.Raw) == 0 {
		t.Error("expected Raw to be populated")
	}
}

func TestGetOrderMarketplaceOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_GB" {
			t.Errorf("marketplace = %q, want EBAY_GB", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"orderId":"o"}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("A"), WithBaseURL(server.URL), WithMarketplace("EBAY_GB"))
	if _, err := c.GetOrder(context.Background(), "o"); err != nil {
		t.Fatalf("GetOrder: %v", err)
	}
}

func TestGetOrderEmptyIDErrors(t *testing.T) {
	c := NewClient(auth.StaticToken("A"))
	if _, err := c.GetOrder(context.Background(), ""); err == nil {
		t.Error("expected error for empty orderID")
	}
}

func TestGetOrderPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1001}]}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("A"), WithBaseURL(server.URL))
	_, err := c.GetOrder(context.Background(), "x")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401, got %v", err)
	}
}

func TestMoneyFloatValue(t *testing.T) {
	cases := []struct {
		in   Money
		want float64
	}{
		{Money{Value: "12.34", Currency: "USD"}, 12.34},
		{Money{Value: "0", Currency: "USD"}, 0},
		{Money{Value: "", Currency: "USD"}, 0},
		{Money{Value: "not-a-number", Currency: "USD"}, 0},
	}
	for _, tc := range cases {
		if got := tc.in.FloatValue(); got != tc.want {
			t.Errorf("FloatValue(%q) = %v, want %v", tc.in.Value, got, tc.want)
		}
	}
}
