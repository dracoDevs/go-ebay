package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

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
	return ParseXMLResponse[GetOrdersResponse](r)
}

type GetOrdersResponse struct {
	BaseResponse
	OrderArray struct {
		Orders []struct {
			OrderID          string `xml:"OrderID"`
			OrderStatus      string `xml:"OrderStatus"`
			AdjustmentAmount Amount `xml:"AdjustmentAmount"`
			AmountPaid       Amount `xml:"AmountPaid"`
			AmountSaved      Amount `xml:"AmountSaved"`
			CheckoutStatus   struct {
				EPaymentStatus                      string `xml:"eBayPaymentStatus"`
				LastModifiedTime                    string `xml:"LastModifiedTime"`
				PaymentMethod                       string `xml:"PaymentMethod"`
				Status                              string `xml:"Status"`
				IntegratedMerchantCreditCardEnabled bool   `xml:"IntegratedMerchantCreditCardEnabled"`
			} `xml:"CheckoutStatus"`
			ShippingDetails struct {
				SalesTax struct {
					ShippingIncludedInTax bool `xml:"ShippingIncludedInTax"`
				} `xml:"SalesTax"`
				ShippingServiceOptions struct {
					ShippingService         string `xml:"ShippingService"`
					ShippingServicePriority int    `xml:"ShippingServicePriority"`
					ExpeditedService        bool   `xml:"ExpeditedService"`
					ShippingTimeMin         int    `xml:"ShippingTimeMin"`
					ShippingTimeMax         int    `xml:"ShippingTimeMax"`
				} `xml:"ShippingServiceOptions"`
				SellingManagerSalesRecordNumber int `xml:"SellingManagerSalesRecordNumber"`
				TaxTable                        struct {
					TaxJurisdiction struct {
						SalesTaxPercent       float64 `xml:"SalesTaxPercent"`
						ShippingIncludedInTax bool    `xml:"ShippingIncludedInTax"`
					} `xml:"TaxJurisdiction"`
				} `xml:"TaxTable"`
			} `xml:"ShippingDetails"`
			CreatedTime     string `xml:"CreatedTime"`
			ShippingAddress struct {
				Name            string `xml:"Name"`
				Street1         string `xml:"Street1"`
				Street2         string `xml:"Street2"`
				CityName        string `xml:"CityName"`
				StateOrProvince string `xml:"StateOrProvince"`
				Country         string `xml:"Country"`
				CountryName     string `xml:"CountryName"`
				Phone           string `xml:"Phone"`
				PostalCode      string `xml:"PostalCode"`
				AddressID       string `xml:"AddressID"`
				AddressOwner    string `xml:"AddressOwner"`
			} `xml:"ShippingAddress"`
			ShippingServiceSelected struct {
				ShippingService     string `xml:"ShippingService"`
				ShippingServiceCost Amount `xml:"ShippingServiceCost"`
			} `xml:"ShippingServiceSelected"`
			Subtotal         Amount `xml:"Subtotal"`
			Total            Amount `xml:"Total"`
			TransactionArray struct {
				Transactions []struct {
					Buyer struct {
						Email         string `xml:"Email"`
						VATStatus     string `xml:"VATStatus"`
						UserFirstName string `xml:"UserFirstName"`
						UserLastName  string `xml:"UserLastName"`
					} `xml:"Buyer"`
					ShippingDetails struct {
						SellingManagerSalesRecordNumber int `xml:"SellingManagerSalesRecordNumber"`
					} `xml:"ShippingDetails"`
					CreatedDate string `xml:"CreatedDate"`
					Item        struct {
						ItemID   string `xml:"ItemID"`
						Location string `xml:"Location"`
						Site     string `xml:"Site"`
						Title    string `xml:"Title"`
					} `xml:"Item"`
					QuantityPurchased int `xml:"QuantityPurchased"`
					Status            struct {
						PaymentHoldStatus string `xml:"PaymentHoldStatus"`
						InquiryStatus     string `xml:"InquiryStatus"`
						ReturnStatus      string `xml:"ReturnStatus"`
					} `xml:"Status"`
					TransactionID           string `xml:"TransactionID"`
					TransactionPrice        Amount `xml:"TransactionPrice"`
					ShippingServiceSelected struct {
						ShippingPackageInfo struct {
							EstimatedDeliveryTimeMin       string `xml:"EstimatedDeliveryTimeMin"`
							EstimatedDeliveryTimeMax       string `xml:"EstimatedDeliveryTimeMax"`
							HandleByTime                   string `xml:"HandleByTime"`
							MinNativeEstimatedDeliveryTime string `xml:"MinNativeEstimatedDeliveryTime"`
							MaxNativeEstimatedDeliveryTime string `xml:"MaxNativeEstimatedDeliveryTime"`
						} `xml:"ShippingPackageInfo"`
					} `xml:"ShippingServiceSelected"`
					ShippedTime              string  `xml:"ShippedTime"`
					FinalValueFee            Amount  `xml:"FinalValueFee"`
					TransactionSiteID        string  `xml:"TransactionSiteID"`
					Platform                 string  `xml:"Platform"`
					ActualShippingCost       Amount  `xml:"ActualShippingCost"`
					ActualHandlingCost       Amount  `xml:"ActualHandlingCost"`
					OrderLineItemID          string  `xml:"OrderLineItemID"`
					InventoryReservationID   string  `xml:"InventoryReservationID"`
					ExtendedOrderID          string  `xml:"ExtendedOrderID"`
					EbayPlusTransaction      BoolStr `xml:"eBayPlusTransaction"`
					GuaranteedShipping       BoolStr `xml:"GuaranteedShipping"`
					EbayCollectAndRemitTax   BoolStr `xml:"eBayCollectAndRemitTax"`
					EbayCollectAndRemitTaxes struct {
						TotalTaxAmount Amount `xml:"TotalTaxAmount"`
						TaxDetails     struct {
							Imposition          string `xml:"Imposition"`
							TaxDescription      string `xml:"TaxDescription"`
							TaxAmount           Amount `xml:"TaxAmount"`
							TaxOnSubtotalAmount Amount `xml:"TaxOnSubtotalAmount"`
							TaxOnShippingAmount Amount `xml:"TaxOnShippingAmount"`
							TaxOnHandlingAmount Amount `xml:"TaxOnHandlingAmount"`
							CollectionMethod    string `xml:"CollectionMethod"`
						} `xml:"TaxDetails"`
					} `xml:"eBayCollectAndRemitTaxes"`
				} `xml:"Transaction"`
			} `xml:"TransactionArray"`
			BuyerUserID        string  `xml:"BuyerUserID"`
			PaidTime           string  `xml:"PaidTime"`
			ShippedTime        string  `xml:"ShippedTime"`
			EIASToken          string  `xml:"EIASToken"`
			PaymentHoldStatus  string  `xml:"PaymentHoldStatus"`
			IsMultiLegShipping BoolStr `xml:"IsMultiLegShipping"`
			MonetaryDetails    struct {
				Payments struct {
					Payment struct {
						PaymentStatus string `xml:"PaymentStatus"`
						Payer         struct {
							Type string `xml:"type,attr"`
							ID   string `xml:",chardata"`
						} `xml:"Payer"`
						Payee struct {
							Type string `xml:"type,attr"`
							ID   string `xml:",chardata"`
						} `xml:"Payee"`
						PaymentTime   string `xml:"PaymentTime"`
						PaymentAmount Amount `xml:"PaymentAmount"`
						ReferenceID   struct {
							Type string `xml:"type,attr"`
							ID   string `xml:",chardata"`
						} `xml:"ReferenceID"`
						FeeOrCreditAmount Amount `xml:"FeeOrCreditAmount"`
					} `xml:"Payment"`
				} `xml:"Payments"`
			} `xml:"MonetaryDetails"`
			SellerUserID                string  `xml:"SellerUserID"`
			SellerEIASToken             string  `xml:"SellerEIASToken"`
			CancelStatus                string  `xml:"CancelStatus"`
			ExtendedOrderID             string  `xml:"ExtendedOrderID"`
			ContainseBayPlusTransaction BoolStr `xml:"ContainseBayPlusTransaction"`
			EbayCollectAndRemitTax      BoolStr `xml:"eBayCollectAndRemitTax"`
		} `xml:"Order"`
	} `xml:"OrderArray"`
}

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
