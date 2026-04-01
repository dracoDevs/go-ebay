package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type GetSellerList struct {
	StartTimeFrom  string      `xml:"StartTimeFrom,omitempty"`
	StartTimeTo    string      `xml:"StartTimeTo,omitempty"`
	EndTimeFrom    string      `xml:"EndTimeFrom,omitempty"`
	EndTimeTo      string      `xml:"EndTimeTo,omitempty"`
	GranularityLevel string    `xml:"GranularityLevel,omitempty"`
	Pagination     *Pagination `xml:"Pagination,omitempty"`
}

func (c GetSellerList) CallName() string { return "GetSellerList" }

func (c GetSellerList) Body() interface{} {
	return c
}

func (c GetSellerList) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetSellerListResponse](r)
}

type GetSellerListResponse struct {
	BaseResponse
	ItemArray        *SellerListItemArray `xml:"ItemArray,omitempty"`
	PaginationResult *PaginationResult    `xml:"PaginationResult,omitempty"`
	HasMoreItems     bool                 `xml:"HasMoreItems"`
	PageNumber       int                  `xml:"PageNumber"`
	ReturnedItemCountActual int           `xml:"ReturnedItemCountActual"`
}

type SellerListItemArray struct {
	Items []SellerListItem `xml:"Item"`
}

type SellerListItem struct {
	ItemID        string         `xml:"ItemID"`
	Title         string         `xml:"Title"`
	SellingStatus *SellingStatus `xml:"SellingStatus,omitempty"`
}
