package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type GetItemTransactions struct {
	ItemID        string `xml:"ItemID"`
	TransactionID string `xml:"TransactionID"`
}

func (c GetItemTransactions) CallName() string {
	return "GetItemTransactions"
}

func (c GetItemTransactions) Body() interface{} {
	return GetItemTransactions{ItemID: c.ItemID, TransactionID: c.TransactionID}
}

func (c GetItemTransactions) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	var xmlResponse GetItemTransactionsResponse
	err := xml.Unmarshal(r, &xmlResponse)

	return xmlResponse, err
}

type GetItemTransactionsResponse struct {
	ebay.OtherEbayResponse

	HasMoreTransactions bool `xml:"HasMoreTransactions,omitempty"`

	Item struct {
		ApplicationData string  `xml:"ApplicationData,omitempty"`
		AutoPay         bool    `xml:"AutoPay,omitempty"`
		BuyItNowPrice   float64 `xml:"BuyItNowPrice,omitempty"`
		CurrencyID      string  `xml:"BuyItNowPrice>currencyID,attr,omitempty"`
		Charity         struct {
			CharityListing bool `xml:"CharityListing,omitempty"`
		} `xml:"Charity,omitempty"`
		Currency                string `xml:"Currency,omitempty"`
		InventoryTrackingMethod string `xml:"InventoryTrackingMethod,omitempty"`
		ItemID                  string `xml:"ItemID,omitempty"`
		ListingDetails          struct {
			EndTime                string `xml:"EndTime,omitempty"`
			HasUnansweredQuestions bool   `xml:"HasUnansweredQuestions,omitempty"`
			StartTime              string `xml:"StartTime,omitempty"`
		} `xml:"ListingDetails,omitempty"`
		ListingType    string `xml:"ListingType,omitempty"`
		Location       string `xml:"Location,omitempty"`
		LotSize        int    `xml:"LotSize,omitempty"`
		PrivateListing bool   `xml:"PrivateListing,omitempty"`
		Quantity       int    `xml:"Quantity,omitempty"`
		Seller         struct {
			EBayGoodStanding        bool    `xml:"eBayGoodStanding,omitempty"`
			EIASToken               string  `xml:"EIASToken,omitempty"`
			Email                   string  `xml:"Email,omitempty"`
			FeedbackPrivate         bool    `xml:"FeedbackPrivate,omitempty"`
			FeedbackScore           int     `xml:"FeedbackScore,omitempty"`
			NewUser                 bool    `xml:"NewUser,omitempty"`
			PositiveFeedbackPercent float64 `xml:"PositiveFeedbackPercent,omitempty"`
			RegistrationDate        string  `xml:"RegistrationDate,omitempty"`
			Site                    string  `xml:"Site,omitempty"`
			Status                  string  `xml:"Status,omitempty"`
			UserID                  string  `xml:"UserID,omitempty"`
			UserIDChanged           bool    `xml:"UserIDChanged,omitempty"`
			VATStatus               string  `xml:"VATStatus,omitempty"`
		} `xml:"Seller,omitempty"`
		SellingStatus struct {
			BidCount              int     `xml:"BidCount,omitempty"`
			ConvertedCurrentPrice float64 `xml:"ConvertedCurrentPrice,omitempty"`
			ConvertedCurrencyID   string  `xml:"ConvertedCurrentPrice>currencyID,attr,omitempty"`
			CurrentPrice          float64 `xml:"CurrentPrice,omitempty"`
			CurrencyID            string  `xml:"CurrentPrice>currencyID,attr,omitempty"`
			FinalValueFee         float64 `xml:"FinalValueFee,omitempty"`
			FeeCurrencyID         string  `xml:"FinalValueFee>currencyID,attr,omitempty"`
			ListingStatus         string  `xml:"ListingStatus,omitempty"`
			QuantitySold          int     `xml:"QuantitySold,omitempty"`
		} `xml:"SellingStatus,omitempty"`
		Site            string  `xml:"Site,omitempty"`
		SKU             string  `xml:"SKU,omitempty"`
		StartPrice      float64 `xml:"StartPrice,omitempty"`
		StartCurrencyID string  `xml:"StartPrice>currencyID,attr,omitempty"`
		Title           string  `xml:"Title,omitempty"`
	} `xml:"Item,omitempty"`

	TransactionArray struct {
		Transactions []struct {
			TransactionID     string `xml:"TransactionID,omitempty"`
			QuantityPurchased int    `xml:"QuantityPurchased,omitempty"`
			Buyer             struct {
				UserID    string `xml:"UserID,omitempty"`
				Email     string `xml:"Email,omitempty"`
				EIASToken string `xml:"EIASToken,omitempty"`
			} `xml:"Buyer,omitempty"`
			AmountPaid      float64 `xml:"AmountPaid,omitempty"`
			CurrencyID      string  `xml:"AmountPaid>currencyID,attr,omitempty"`
			OrderLineItemID string  `xml:"OrderLineItemID,omitempty"`

			CreatedDate string `xml:"CreatedDate,omitempty"`
			PaidTime    string `xml:"PaidTime,omitempty"`
			ShippedTime string `xml:"ShippedTime,omitempty"`
			Status      struct {
				CompleteStatus        string `xml:"CompleteStatus,omitempty"`
				CheckoutStatus        string `xml:"CheckoutStatus,omitempty"`
				PaymentMethodUsed     string `xml:"PaymentMethodUsed,omitempty"`
				PaymentHoldStatus     string `xml:"PaymentHoldStatus,omitempty"`
				BuyerSelectedShipping bool   `xml:"BuyerSelectedShipping,omitempty"`
				LastTimeModified      string `xml:"LastTimeModified,omitempty"`
				PaymentInstrument     string `xml:"PaymentInstrument,omitempty"`
				RefundStatus          string `xml:"RefundStatus,omitempty"`
				ReturnStatus          string `xml:"ReturnStatus,omitempty"`
			} `xml:"Status,omitempty"`
			ShippingDetails struct {
				ShipmentTrackingDetails []struct {
					ShipmentTrackingNumber string `xml:"ShipmentTrackingNumber,omitempty"`
					ShippingCarrierUsed    string `xml:"ShippingCarrierUsed,omitempty"`
				} `xml:"ShipmentTrackingDetails,omitempty"`
			} `xml:"ShippingDetails,omitempty"`
			Variation struct {
				SKU string `xml:"SKU,omitempty"`
			} `xml:"Variation,omitempty"`
			// Add more fields as needed from the full schema
		} `xml:"Transaction,omitempty"`
	} `xml:"TransactionArray,omitempty"`
}

func (r GetItemTransactionsResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}
