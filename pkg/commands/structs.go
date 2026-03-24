package commands

import "encoding/xml"

type Amount struct {
	Value      float64 `xml:",chardata"`
	CurrencyID string  `xml:"currencyID,attr"`
}

type BoolStr bool

func (b *BoolStr) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	if err := d.DecodeElement(&v, &start); err != nil {
		return err
	}
	*b = (v == "true")
	return nil
}

type Pagination struct {
	EntriesPerPage int `xml:"EntriesPerPage"`
	PageNumber     int `xml:"PageNumber"`
}

type Storefront struct {
	StoreCategoryID string
}

type ReturnPolicy struct {
	ReturnsAccepted, ReturnsAcceptedOption, ReturnsWithinOption, RefundOption, ShippingCostPaidByOption string
}

type ItemSpecifics struct {
	NameValueList []NameValueList
}

type NameValueList struct {
	Name  string
	Value string
}

type PictureDetails struct {
	PictureURL []string
	GalleryURL string `xml:",omitempty"`
}

type PrimaryCategory struct {
	CategoryID string
}

type BestOfferDetails struct {
	BestOfferEnabled bool
}

type BrandMPN struct {
	Brand, MPN string
}

type ProductListingDetails struct {
	UPC string
	// BrandMPN BrandMPN
}

type ShippingDetails struct {
	ShippingType                           string
	ShippingDiscountProfileID              string
	InternationalShippingDiscountProfileID string
	ExcludeShipToLocation                  []string
	ShippingServiceOptions                 []ShippingServiceOption
	InternationalShippingServiceOption     []InternationalShippingServiceOption
}

type ShippingServiceOption struct {
	ShippingService               string
	ShippingServiceCost           float64
	ShippingServiceAdditionalCost float64
	ShippingServicePriority       int
	FreeShipping                  bool
}

type InternationalShippingServiceOption struct {
	ShippingService               string
	ShippingServiceCost           float64
	ShippingServiceAdditionalCost float64
	ShipToLocation                []string
	ShippingServicePriority       int
}
