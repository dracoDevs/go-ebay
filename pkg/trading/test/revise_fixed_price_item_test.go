package test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/trading"
)

func TestReviseFixedPriceItemSKUOmitWhenEmpty(t *testing.T) {
	cmd := trading.ReviseFixedPriceItem{ItemID: "123", Quantity: trading.UintPtr(5)}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(out), "<SKU>") {
		t.Errorf("expected SKU to be omitted; got %s", out)
	}
}

func TestReviseFixedPriceItemSKUEmittedWhenSet(t *testing.T) {
	cmd := trading.ReviseFixedPriceItem{ItemID: "123", SKU: "abc-sku"}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !strings.Contains(string(out), "<SKU>abc-sku</SKU>") {
		t.Errorf("expected <SKU>abc-sku</SKU>; got %s", out)
	}
}

func TestReviseFixedPriceItemQuantityOmittedWhenNil(t *testing.T) {
	cmd := trading.ReviseFixedPriceItem{ItemID: "123", StartPrice: "9.99"}
	out, err := xml.Marshal(cmd.Body())
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.Contains(string(out), "<Quantity>") {
		t.Errorf("expected Quantity omitted when nil; got %s", out)
	}
}

func TestReviseFixedPriceItemQuantityEmittedWhenSet(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		cmd := trading.ReviseFixedPriceItem{ItemID: "123", Quantity: trading.UintPtr(0)}
		out, _ := xml.Marshal(cmd.Body())
		if !strings.Contains(string(out), "<Quantity>0</Quantity>") {
			t.Errorf("expected <Quantity>0</Quantity>; got %s", out)
		}
	})
	t.Run("nonzero", func(t *testing.T) {
		cmd := trading.ReviseFixedPriceItem{ItemID: "123", Quantity: trading.UintPtr(7)}
		out, _ := xml.Marshal(cmd.Body())
		if !strings.Contains(string(out), "<Quantity>7</Quantity>") {
			t.Errorf("expected <Quantity>7</Quantity>; got %s", out)
		}
	})
}

func TestReviseFixedPriceItemResponseParsesRelistedItemID(t *testing.T) {
	xmlBody := []byte(`
<ReviseFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
	<Ack>Success</Ack>
	<ItemID>OLDID</ItemID>
	<RelistedItemID>NEWID</RelistedItemID>
</ReviseFixedPriceItemResponse>`)
	resp, err := trading.ReviseFixedPriceItem{}.ParseResponse(xmlBody)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	r := resp.(trading.ReviseFixedPriceItemResponse)
	if r.ItemID != "OLDID" {
		t.Errorf("ItemID = %q", r.ItemID)
	}
	if r.RelistedItemID != "NEWID" {
		t.Errorf("RelistedItemID = %q", r.RelistedItemID)
	}
}

func TestReviseFixedPriceItemResponseRelistedItemIDEmptyWhenAbsent(t *testing.T) {
	xmlBody := []byte(`
<ReviseFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
	<Ack>Success</Ack>
	<ItemID>SAMEID</ItemID>
</ReviseFixedPriceItemResponse>`)
	resp, err := trading.ReviseFixedPriceItem{}.ParseResponse(xmlBody)
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	r := resp.(trading.ReviseFixedPriceItemResponse)
	if r.RelistedItemID != "" {
		t.Errorf("RelistedItemID = %q, want empty", r.RelistedItemID)
	}
}
