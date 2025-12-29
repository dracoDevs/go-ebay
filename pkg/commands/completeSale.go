package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type CompleteSale struct {
	ItemID        string    `xml:"ItemID,omitempty"`
	Shipped       *bool     `xml:"Shipped,omitempty"`
	TransactionID string    `xml:"TransactionID,omitempty"`
	Shipment      *Shipment `xml:"Shipment,omitempty"`
}

type Shipment struct {
	ShipmentTrackingDetails ShipmentTrackingDetails `xml:"ShipmentTrackingDetails"`
}

type ShipmentTrackingDetails struct {
	ShipmentTrackingNumber string `xml:"ShipmentTrackingNumber"`
	ShippingCarrierUsed    string `xml:"ShippingCarrierUsed"`
}

func (c CompleteSale) CallName() string {
	return "CompleteSale"
}

func (c CompleteSale) Body() interface{} {
	return CompleteSale{
		ItemID:        c.ItemID,
		Shipped:       c.Shipped,
		TransactionID: c.TransactionID,
		Shipment:      c.Shipment,
	}
}

func (c CompleteSale) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	var xmlResponse CompleteSaleResponse
	err := xml.Unmarshal(r, &xmlResponse)
	return xmlResponse, err
}

type CompleteSaleResponse struct {
	ebay.OtherEbayResponse
}

func (r CompleteSaleResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}
