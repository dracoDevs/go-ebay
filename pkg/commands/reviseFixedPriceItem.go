package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type ReviseFixedPriceItemResponse struct {
	ebay.OtherEbayResponse
}

type ReviseFixedPriceItem struct {
	ItemID                string
	StartPrice            string `xml:",omitempty"`
	ConditionID           uint   `xml:",omitempty"`
	Quantity              uint
	Title                 string           `xml:",omitempty"`
	Description           string           `xml:",omitempty"`
	PayPalEmailAddress    string           `xml:",omitempty"`
	PictureDetails        *PictureDetails  `xml:",omitempty"`
	ShippingDetails       *ShippingDetails `xml:",omitempty"`
	PrimaryCategory       *PrimaryCategory
	ReturnPolicy          *ReturnPolicy          `xml:",omitempty"`
	ProductListingDetails *ProductListingDetails `xml:",omitempty"`
	ItemSpecifics         map[string]string      `xml:",omitempty"`
}

func (c ReviseFixedPriceItem) CallName() string {
	return "ReviseFixedPriceItem"
}

func (c ReviseFixedPriceItem) Body() interface{} {
	type Item struct {
		ReviseFixedPriceItem
	}

	return Item{c}
}

func (c ReviseFixedPriceItem) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	var xmlResponse ReviseFixedPriceItemResponse
	err := xml.Unmarshal(r, &xmlResponse)

	return xmlResponse, err
}

func (r ReviseFixedPriceItemResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}
