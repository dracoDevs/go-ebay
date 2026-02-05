package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type AddFixedPriceItem struct {
	Currency              string
	Country               string
	DispatchTimeMax       int    `xml:",omitempty"`
	ConditionID           int    `xml:",omitempty"`
	Title                 string `xml:",omitempty"`
	Description           string `xml:",omitempty"`
	StartPrice            string
	ListingType           string            `xml:",omitempty"`
	Quantity              uint              `xml:",omitempty"`
	BestOfferDetails      *BestOfferDetails `xml:",omitempty"`
	PaymentMethods        string            `xml:",omitempty"`
	PayPalEmailAddress    string            `xml:",omitempty"`
	ListingDuration       string
	ShippingDetails       *ShippingDetails `xml:",omitempty"`
	PrimaryCategory       *PrimaryCategory
	Storefront            *Storefront            `xml:",omitempty"`
	PostalCode            string                 `xml:",omitempty"`
	ReturnPolicy          *ReturnPolicy          `xml:",omitempty"`
	PictureDetails        *PictureDetails        `xml:",omitempty"`
	ProductListingDetails *ProductListingDetails `xml:",omitempty"`
	ItemSpecifics         ItemSpecifics          `xml:",omitempty"`
}

func (c AddFixedPriceItem) CallName() string { return "AddFixedPriceItem" }

func (c AddFixedPriceItem) Body() interface{} {
	type Item struct{ AddFixedPriceItem }
	return Item{c}
}

func (c AddFixedPriceItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[AddFixedPriceItemResponse](r)
}

type AddFixedPriceItemResponse struct {
	BaseResponse
	ItemID string
}
