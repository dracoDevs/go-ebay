package utils

import (
	"strings"
	"testing"
)

func TestRemoveTagXML(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		tag      string
		contains string
		excludes string
	}{
		{
			name:     "removes wrapping tag",
			xml:      `<Root><GetOrders><OrderID>123</OrderID></GetOrders></Root>`,
			tag:      "GetOrders",
			contains: "<OrderID>123</OrderID>",
			excludes: "<GetOrders>",
		},
		{
			name:     "no matching tag leaves XML unchanged",
			xml:      `<Root><Data>test</Data></Root>`,
			tag:      "Missing",
			contains: "<Data>test</Data>",
		},
		{
			name:     "removes EndItem tag",
			xml:      `<EndItemRequest><EndItem><ItemID>456</ItemID></EndItem></EndItemRequest>`,
			tag:      "EndItem",
			contains: "<ItemID>456</ItemID>",
			excludes: "<EndItem>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveTagXML(tt.xml, tt.tag)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("result %q does not contain %q", result, tt.contains)
			}
			if tt.excludes != "" && strings.Contains(result, tt.excludes) {
				t.Errorf("result %q should not contain %q", result, tt.excludes)
			}
		})
	}
}
