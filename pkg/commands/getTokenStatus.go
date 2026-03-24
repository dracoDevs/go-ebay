package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type TokenStatusResp struct {
	Status         string `xml:"Status,omitempty"`
	EIASToken      string `xml:"EIASToken,omitempty"`
	ExpirationTime string `xml:"ExpirationTime,omitempty"`
	RevocationTime string `xml:"RevocationTime,omitempty"`
}

type GetTokenStatus struct{}

func (c GetTokenStatus) CallName() string { return "GetTokenStatus" }

func (c GetTokenStatus) Body() interface{} {
	type body struct{}
	return body{}
}

func (c GetTokenStatus) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetTokenStatusResponse](r)
}

type GetTokenStatusResponse struct {
	BaseResponse
	Version               string           `xml:"Version,omitempty"`
	Build                 string           `xml:"Build,omitempty"`
	TokenStatus           *TokenStatusResp `xml:"TokenStatus,omitempty"`
	HardExpirationWarning string           `xml:"HardExpirationWarning,omitempty"`
}
