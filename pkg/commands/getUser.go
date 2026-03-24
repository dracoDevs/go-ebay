package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type GetUser struct{}

func (c GetUser) CallName() string { return "GetUser" }

func (c GetUser) Body() interface{} {
	type body struct{}
	return body{}
}

func (c GetUser) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetUserResponse](r)
}

type GetUserResponse struct {
	BaseResponse
	User struct {
		UserID    string `xml:"UserID"`
		EIASToken string `xml:"EIASToken"`
		Email     string `xml:"Email"`
		Status    string `xml:"Status"`
	} `xml:"User"`
}
