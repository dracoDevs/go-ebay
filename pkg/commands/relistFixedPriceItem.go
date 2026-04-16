package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type RelistFixedPriceItem struct {
	ItemID   string
	Quantity uint `xml:",omitempty"`
}

func (c RelistFixedPriceItem) CallName() string { return "RelistFixedPriceItem" }

func (c RelistFixedPriceItem) Body() interface{} {
	type Item struct{ RelistFixedPriceItem }
	return Item{c}
}

func (c RelistFixedPriceItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[RelistFixedPriceItemResponse](r)
}

type RelistFixedPriceItemResponse struct {
	BaseResponse
	ItemID string `xml:"ItemID"`
}
