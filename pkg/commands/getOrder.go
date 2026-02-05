package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type OrderIDArray struct {
	OrderID string `xml:"OrderID"`
}

type GetOrders struct {
	NumberOfDays         int
	IncludeFinalValueFee bool
	OrderIDArray         OrderIDArray
}

func (c GetOrders) CallName() string { return "GetOrders" }

func (c GetOrders) Body() interface{} {
	return c
}

func (c GetOrders) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetOrderResponse](r)
}

type GetOrderResponse struct {
	BaseResponse
	Item struct {
		ItemID         string
		Quantity       int64
		ListingDetails struct {
			StartTime string
		}
		SellingStatus struct {
			ListingStatus string
			QuantitySold  int64
			CurrentPrice  float64
		}
	}
	OrderArray struct {
		Orders []struct {
			OrderID         string `xml:"OrderID"`
			BuyerUserID     string `xml:"BuyerUserID"`
			ShippingAddress struct {
				Name            string `xml:"Name"`
				Street1         string `xml:"Street1"`
				Street2         string `xml:"Street2"`
				CityName        string `xml:"CityName"`
				StateOrProvince string `xml:"StateOrProvince"`
				Country         string `xml:"Country"`
				CountryName     string `xml:"CountryName"`
				PostalCode      string `xml:"PostalCode"`
				Phone           string `xml:"Phone"`
			} `xml:"ShippingAddress"`
		} `xml:"Order"`
	} `xml:"OrderArray"`
}
