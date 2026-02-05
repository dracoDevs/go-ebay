package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type AddItem struct {
	Currency              string
	Country               string
	DispatchTimeMax       int    `xml:",omitempty"`
	ConditionID           int    `xml:",omitempty"`
	Title                 string `xml:",omitempty"`
	Description           string `xml:",omitempty"`
	StartPrice            string
	BuyItNowPrice         string `xml:",omitempty"`
	ListingType           string `xml:",omitempty"`
	Quantity              uint   `xml:",omitempty"`
	PaymentMethods        string `xml:",omitempty"`
	PayPalEmailAddress    string `xml:",omitempty"`
	ListingDuration       string
	ShippingDetails       *ShippingDetails `xml:",omitempty"`
	PrimaryCategory       *PrimaryCategory
	Storefront            *Storefront            `xml:",omitempty"`
	PostalCode            string                 `xml:",omitempty"`
	ReturnPolicy          *ReturnPolicy          `xml:",omitempty"`
	PictureDetails        *PictureDetails        `xml:",omitempty"`
	ProductListingDetails *ProductListingDetails `xml:",omitempty"`
	ItemSpecifics         []ItemSpecifics        `xml:",omitempty"`
}

func (c AddItem) CallName() string { return "AddItem" }

func (c AddItem) Body() interface{} {
	type Item struct{ AddItem }
	return Item{c}
}

func (c AddItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[AddItemResponse](r)
}

type AddItemResponse struct {
	BaseResponse
	ItemID string
}
