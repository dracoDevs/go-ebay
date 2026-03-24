package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type GetMyeBaySelling struct {
	ActiveList *ActiveListRequest `xml:"ActiveList,omitempty"`
}

type ActiveListRequest struct {
	Sort       string      `xml:"Sort,omitempty"`
	Pagination *Pagination `xml:"Pagination,omitempty"`
}

func (c GetMyeBaySelling) CallName() string { return "GetMyeBaySelling" }

func (c GetMyeBaySelling) Body() interface{} {
	return c
}

func (c GetMyeBaySelling) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetMyeBaySellingResponse](r)
}

type GetMyeBaySellingResponse struct {
	BaseResponse
	ActiveList *ActiveListResult `xml:"ActiveList,omitempty"`
}

type ActiveListResult struct {
	ItemArray        *SellingItemArray `xml:"ItemArray,omitempty"`
	PaginationResult *PaginationResult `xml:"PaginationResult,omitempty"`
}

type SellingItemArray struct {
	Items []SellingItem `xml:"Item"`
}

type SellingItem struct {
	ItemID            string          `xml:"ItemID"`
	Title             string          `xml:"Title"`
	PictureDetails    *PictureDetails `xml:"PictureDetails,omitempty"`
	ListingDetails    *ListingDetails `xml:"ListingDetails,omitempty"`
	SellingStatus     *SellingStatus  `xml:"SellingStatus,omitempty"`
	Quantity          int             `xml:"Quantity"`
	QuantityAvailable int             `xml:"QuantityAvailable"`
}

type ListingDetails struct {
	StartTime   string `xml:"StartTime"`
	ViewItemURL string `xml:"ViewItemURL"`
}

type SellingStatus struct {
	ListingStatus string  `xml:"ListingStatus"`
	CurrentPrice  float64 `xml:"CurrentPrice"`
	QuantitySold  int     `xml:"QuantitySold"`
}

type PaginationResult struct {
	TotalNumberOfPages   int `xml:"TotalNumberOfPages"`
	TotalNumberOfEntries int `xml:"TotalNumberOfEntries"`
}
