package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type TokenStatusResp struct {
	Status         string `xml:"Status,omitempty"`
	EIASToken      string `xml:"EIASToken,omitempty"`
	ExpirationTime string `xml:"ExpirationTime,omitempty"`
	RevocationTime string `xml:"RevocationTime,omitempty"`
}

type GetTokenStatus struct {
	XMLName       xml.Name `xml:"GetTokenStatusRequest"`
	Xmlns         string   `xml:"xmlns,attr,omitempty"`
	ErrorLanguage string   `xml:"ErrorLanguage,omitempty"`
	MessageID     string   `xml:"MessageID,omitempty"`
	Version       string   `xml:"Version,omitempty"`
	WarningLevel  string   `xml:"WarningLevel,omitempty"`
}

func (c GetTokenStatus) CallName() string { return "GetTokenStatus" }

func (c GetTokenStatus) Body() interface{} {
	out := c
	if out.Xmlns == "" {
		out.Xmlns = "urn:ebay:apis:eBLBaseComponents"
	}
	return out
}

func (c GetTokenStatus) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetTokenStatusResponse](r)
}

type GetTokenStatusResponse struct {
	XMLName xml.Name `xml:"GetTokenStatusResponse"`
	BaseResponse
	Timestamp             string           `xml:"Timestamp,omitempty"`
	Ack                   string           `xml:"Ack,omitempty"`
	Version               string           `xml:"Version,omitempty"`
	Build                 string           `xml:"Build,omitempty"`
	TokenStatus           *TokenStatusResp `xml:"TokenStatus,omitempty"`
	HardExpirationWarning string           `xml:"HardExpirationWarning,omitempty"`
}
