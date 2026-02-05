package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type GetItem struct {
	ItemID string
}

func (c GetItem) CallName() string { return "GetItem" }

func (c GetItem) Body() interface{} {
	type ItemID struct {
		ItemID string `xml:",innerxml"`
	}
	return ItemID{c.ItemID}
}

func (c GetItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetItemResponse](r)
}

type GetItemResponse struct {
	BaseResponse
	Item struct {
		ItemID         string
		Quantity       int64
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
