package trading

import (
	"encoding/xml"
	"fmt"
	"time"
)

type request struct {
	conf    Conf
	command Command
}

type Response interface {
	Failure() bool
	ResponseErrors() Errors
}

type GenericResponse struct {
	Timestamp Timestamp
	Ack       string
	Errors    []responseError
}

type responseError struct {
	ShortMessage        string
	LongMessage         string
	ErrorCode           int
	SeverityCode        string
	ErrorClassification string
}

func (r GenericResponse) Failure() bool {
	return r.Ack == "Failure"
}

func (r GenericResponse) ResponseErrors() Errors {
	return r.Errors
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse time %q", raw)
}

func (c request) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	startElement := xml.StartElement{
		Name: xml.Name{
			Space: "urn:ebay:apis:eBLBaseComponents",
			Local: fmt.Sprintf("%sRequest", c.command.CallName()),
		},
	}

	err := e.EncodeToken(startElement)

	if err != nil {
		return err
	}

	type RequesterCredentials struct {
		EbayAuthToken string `xml:"eBayAuthToken"`
	}

	err = e.Encode(
		RequesterCredentials{
			EbayAuthToken: c.conf.AuthToken,
		},
	)

	if err != nil {
		return err
	}

	err = e.Encode(c.command.Body())

	if err != nil {
		return err
	}

	endElement := xml.EndElement{
		Name: xml.Name{
			Space: "urn:ebay:apis:eBLBaseComponents",
			Local: fmt.Sprintf("%sRequest", c.command.CallName()),
		},
	}

	err = e.EncodeToken(endElement)

	if err != nil {
		return err
	}

	return nil
}
