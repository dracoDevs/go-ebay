package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/identity"
)

func TestGetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"userId":"U1","username":"seller1","accountType":"INDIVIDUAL","registrationMarketplaceId":"EBAY_US"}`))
	}))
	defer server.Close()

	c := identity.NewClient(auth.StaticToken("A"), identity.WithBaseURL(server.URL))
	u, err := c.GetUser(context.Background())
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.Username != "seller1" || u.AccountType != "INDIVIDUAL" {
		t.Errorf("user = %+v", u)
	}
}

func TestGetUserPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1001}]}`))
	}))
	defer server.Close()

	c := identity.NewClient(auth.StaticToken("A"), identity.WithBaseURL(server.URL))
	_, err := c.GetUser(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401, got %v", err)
	}
}
