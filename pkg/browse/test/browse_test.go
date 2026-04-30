package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/browse"
)

func TestGetItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/item/v1|123|0" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"itemId":"v1|123|0","legacyItemId":"123","title":"Item","price":{"value":"9.99","currency":"USD"},"image":{"imageUrl":"https://i.ebay/x.jpg"}}`))
	}))
	defer server.Close()

	c := browse.NewClient(auth.StaticToken("A"), browse.WithBaseURL(server.URL))
	out, err := c.GetItemByLegacyID(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetItemByLegacyID: %v", err)
	}
	if out.LegacyItemID != "123" || out.Title != "Item" || out.Price.Value != "9.99" {
		t.Errorf("item = %+v", out)
	}
	if len(out.Raw) == 0 {
		t.Error("expected Raw populated")
	}
}

func TestGetItemRequiresID(t *testing.T) {
	c := browse.NewClient(auth.StaticToken("A"))
	_, err := c.GetItem(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "itemID is required") {
		t.Errorf("expected itemID error, got %v", err)
	}
}
