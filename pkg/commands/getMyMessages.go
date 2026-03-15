package commands

import "github.com/dracoDevs/go-ebay/pkg/ebay"

type GetMyMessages struct {
	MessageIDs  MessageIDs
	DetailLevel string
}

type MessageIDs struct {
	MessageID string `xml:"MessageID"`
}

func (c GetMyMessages) CallName() string { return "GetMyMessages" }

func (c GetMyMessages) Body() interface{} {
	return c
}

func (c GetMyMessages) ParseResponse(r []byte) (ebay.EbayResponse, error) {
	return ParseXMLResponse[GetMyMessagesResponse](r)
}

type GetMyMessagesResponse struct {
	BaseResponse
	Messages struct {
		Message []MyMessage `xml:"Message"`
	} `xml:"Messages"`
}

type MyMessage struct {
	MessageID   string `xml:"MessageID"`
	ItemID      string `xml:"ItemID"`
	Subject     string `xml:"Subject"`
	Sender      string `xml:"Sender"`
	MessageType string `xml:"MessageType"`
	Text        string `xml:"Text"`
}
