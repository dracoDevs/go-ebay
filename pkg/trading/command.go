package trading

import (
	"encoding/xml"
)

type BaseResponse struct {
	GenericResponse
}

func (r BaseResponse) ResponseErrors() Errors {
	return r.GenericResponse.Errors
}

func ParseXMLResponse[T Response](r []byte) (Response, error) {
	var resp T
	err := xml.Unmarshal(r, &resp)
	return resp, err
}
