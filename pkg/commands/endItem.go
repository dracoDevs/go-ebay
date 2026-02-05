package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type EndItem struct {
	ItemID       string       `xml:"ItemID"`
	EndingReason EndingReason `xml:"EndingReason"`
}

type EndingReason string

const (
	CustomCode        EndingReason = "CustomCode"
	Incorrect         EndingReason = "Incorrect"
	LostOrBroken      EndingReason = "LostOrBroken"
	NotAvailable      EndingReason = "NotAvailable"
	OtherListingError EndingReason = "OtherListingError"
	ProductDeleted    EndingReason = "ProductDeleted"
	SellToHighBidder  EndingReason = "SellToHighBidder"
	Sold              EndingReason = "Sold"
)

func (c EndItem) CallName() string { return "EndItem" }

func (c EndItem) Body() interface{} { return c }

func (c EndItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[EndItemResponse](r)
}

type EndItemResponse struct {
	BaseResponse
	EndTime string `xml:"EndTime"`
}
