package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/account"
	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestListPaymentPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment_policy" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("marketplace_id"); got != "EBAY_US" {
			t.Errorf("marketplace_id = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paymentPolicies":[{"paymentPolicyId":"P1","name":"Default","marketplaceId":"EBAY_US"}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("ACCESS"), account.WithBaseURL(server.URL))
	out, err := c.ListPaymentPolicies(context.Background(), "EBAY_US")
	if err != nil {
		t.Fatalf("ListPaymentPolicies: %v", err)
	}
	if len(out) != 1 || out[0].PaymentPolicyID != "P1" {
		t.Errorf("out = %+v", out)
	}
}

func TestCreatePaymentPolicyParsesLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		w.Header().Set("Location", "/sell/account/v1/payment_policy/P-NEW-123")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("ACCESS"), account.WithBaseURL(server.URL))
	id, err := c.CreatePaymentPolicy(context.Background(), account.PaymentPolicy{
		Name:          "Default",
		MarketplaceID: "EBAY_US",
		CategoryTypes: []account.CategoryType{{Name: "ALL_EXCLUDING_MOTORS_VEHICLES"}},
		ImmediatePay:  true,
	})
	if err != nil {
		t.Fatalf("CreatePaymentPolicy: %v", err)
	}
	if id != "P-NEW-123" {
		t.Errorf("id = %q", id)
	}
}

func TestCreateReturnPolicyFromResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"returnPolicyId":"R-99","name":"30-day"}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	id, err := c.CreateReturnPolicy(context.Background(), account.ReturnPolicy{
		Name:            "30-day",
		MarketplaceID:   "EBAY_US",
		CategoryTypes:   []account.CategoryType{{Name: "ALL_EXCLUDING_MOTORS_VEHICLES"}},
		ReturnsAccepted: true,
		ReturnPeriod:    &account.TimeDuration{Unit: "DAY", Value: 30},
	})
	if err != nil {
		t.Fatalf("CreateReturnPolicy: %v", err)
	}
	if id != "R-99" {
		t.Errorf("id = %q", id)
	}
}

func TestListFulfillmentPolicies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fulfillment_policy" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"fulfillmentPolicies":[{"fulfillmentPolicyId":"F1","name":"Standard"}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	out, err := c.ListFulfillmentPolicies(context.Background(), "EBAY_US")
	if err != nil {
		t.Fatalf("ListFulfillmentPolicies: %v", err)
	}
	if len(out) != 1 || out[0].FulfillmentPolicyID != "F1" {
		t.Errorf("out = %+v", out)
	}
}

func TestPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1001}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	_, err := c.ListPaymentPolicies(context.Background(), "EBAY_US")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401, got %v", err)
	}
}
