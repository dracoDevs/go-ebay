package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type CompleteSale struct {
	ItemID          *string   `xml:"ItemID,omitempty"`
	Shipped         *bool     `xml:"Shipped,omitempty"`
	TransactionID   *string   `xml:"TransactionID,omitempty"`
	Shipment        *Shipment `xml:"Shipment,omitempty"`
	OrderLineItemID *string   `xml:"OrderLineItemID,omitempty"`
}

type Shipment struct {
	ShipmentTrackingDetails ShipmentTrackingDetails `xml:"ShipmentTrackingDetails"`
}

type ShipmentTrackingDetails struct {
	ShipmentTrackingNumber string `xml:"ShipmentTrackingNumber"`
	ShippingCarrierUsed    string `xml:"ShippingCarrierUsed"`
}

func (c CompleteSale) CallName() string { return "CompleteSale" }

func (c CompleteSale) Body() interface{} { return c }

func (c CompleteSale) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[CompleteSaleResponse](r)
}

type CompleteSaleResponse struct {
	BaseResponse
}
