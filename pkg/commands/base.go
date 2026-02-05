package commands

import (
	"encoding/xml"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

type BaseResponse struct {
	ebay.OtherEbayResponse
}

func (r BaseResponse) ResponseErrors() ebay.EbayErrors {
	return r.OtherEbayResponse.Errors
}

func ParseXMLResponse[T ebay.EbayResponse](r []byte) (ebay.EbayResponse, error) {
	var resp T
	err := xml.Unmarshal(r, &resp)
	return resp, err
}
