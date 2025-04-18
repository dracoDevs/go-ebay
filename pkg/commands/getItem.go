package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type GetItem struct {
	ItemID string
}

func (c GetItem) CallName() string {
	return "GetItem"
}

func (c GetItem) Body() interface{} {
	type ItemID struct {
		ItemID string `xml:",innerxml"`
	}

	return ItemID{c.ItemID}
}

func (c GetItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	var xmlResponse GetItemResponse
	err := xml.Unmarshal(r, &xmlResponse)

	return xmlResponse, err
}

type GetItemResponse struct {
	ebay.OtherEbayResponse

	Item struct {
		ItemID        string
		Quantity      int64
		ListingDetails struct {
			StartTime string
		}
		SellingStatus struct {
			ListingStatus string
			QuantitySold  int64
			CurrentPrice  float64
		}
	}
}

func (r GetItemResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}
