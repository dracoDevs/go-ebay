package trading

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestReviseFixedPriceItemSKUOmitWhenEmpty confirms that a zero-valued SKU
// is omitted from the request body. Without this, every existing caller
// (which never set SKU) would start emitting <SKU></SKU> and could collide
// with eBay's own SKU validation.
func TestReviseFixedPriceItemSKUOmitWhenEmpty(t *testing.T) {
	cmd := ReviseFixedPriceItem{ItemID: "123", Quantity: UintPtr(5)}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(out), "<SKU>") {
		t.Errorf("expected SKU element to be omitted; got %s", out)
	}
}

// TestReviseFixedPriceItemSKUEmittedWhenSet confirms the new SKU field
// reaches the wire when set. This is the path used by the lazy-migrate
// flow that prepares a Trading-born listing for the Inventory API.
func TestReviseFixedPriceItemSKUEmittedWhenSet(t *testing.T) {
	cmd := ReviseFixedPriceItem{ItemID: "123", SKU: "abc-sku"}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), "<SKU>abc-sku</SKU>") {
		t.Errorf("expected <SKU>abc-sku</SKU>; got %s", out)
	}
}

// TestReviseFixedPriceItemQuantityOmittedWhenNil is the latent-bug fix:
// a price-only or SKU-only revise must not silently zero the listing's
// quantity by emitting <Quantity>0</Quantity>.
func TestReviseFixedPriceItemQuantityOmittedWhenNil(t *testing.T) {
	cmd := ReviseFixedPriceItem{ItemID: "123", StartPrice: "9.99"}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(out), "<Quantity>") {
		t.Errorf("expected Quantity element to be omitted when nil; got %s", out)
	}
}

// TestReviseFixedPriceItemQuantityEmittedWhenSet covers both Quantity=0
// (legitimate "set to zero" via pointer) and Quantity=N.
func TestReviseFixedPriceItemQuantityEmittedWhenSet(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		cmd := ReviseFixedPriceItem{ItemID: "123", Quantity: UintPtr(0)}
		out, _ := xml.Marshal(cmd.Body())
		if !strings.Contains(string(out), "<Quantity>0</Quantity>") {
			t.Errorf("expected <Quantity>0</Quantity>; got %s", out)
		}
	})
	t.Run("nonzero", func(t *testing.T) {
		cmd := ReviseFixedPriceItem{ItemID: "123", Quantity: UintPtr(7)}
		out, _ := xml.Marshal(cmd.Body())
		if !strings.Contains(string(out), "<Quantity>7</Quantity>") {
			t.Errorf("expected <Quantity>7</Quantity>; got %s", out)
		}
	})
}

// TestReviseFixedPriceItemResponseParsesRelistedItemID confirms the response
// surfaces RelistedItemID, which eBay returns when the revise was satisfied
// by minting a new listing instead of editing the original.
func TestReviseFixedPriceItemResponseParsesRelistedItemID(t *testing.T) {
	xmlBody := []byte(`
<ReviseFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
	<Ack>Success</Ack>
	<ItemID>OLDID</ItemID>
	<RelistedItemID>NEWID</RelistedItemID>
</ReviseFixedPriceItemResponse>`)
	resp, err := ParseXMLResponse[ReviseFixedPriceItemResponse](xmlBody)
	if err != nil {
		t.Fatalf("ParseXMLResponse: %v", err)
	}
	r := resp.(ReviseFixedPriceItemResponse)
	if r.ItemID != "OLDID" {
		t.Errorf("ItemID = %q, want OLDID", r.ItemID)
	}
	if r.RelistedItemID != "NEWID" {
		t.Errorf("RelistedItemID = %q, want NEWID", r.RelistedItemID)
	}
}

// TestReviseFixedPriceItemResponseRelistedItemIDEmptyWhenAbsent confirms the
// common case (no relist happened) leaves the field as the zero value.
func TestReviseFixedPriceItemResponseRelistedItemIDEmptyWhenAbsent(t *testing.T) {
	xmlBody := []byte(`
<ReviseFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
	<Ack>Success</Ack>
	<ItemID>SAMEID</ItemID>
</ReviseFixedPriceItemResponse>`)
	resp, err := ParseXMLResponse[ReviseFixedPriceItemResponse](xmlBody)
	if err != nil {
		t.Fatalf("ParseXMLResponse: %v", err)
	}
	r := resp.(ReviseFixedPriceItemResponse)
	if r.RelistedItemID != "" {
		t.Errorf("RelistedItemID = %q, want empty", r.RelistedItemID)
	}
}
