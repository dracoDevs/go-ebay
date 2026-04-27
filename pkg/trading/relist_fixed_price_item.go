package trading

type RelistFixedPriceItem struct {
	ItemID   string
	Quantity uint `xml:",omitempty"`
}

func (c RelistFixedPriceItem) CallName() string { return "RelistFixedPriceItem" }

func (c RelistFixedPriceItem) Body() interface{} {
	type Item struct{ RelistFixedPriceItem }
	return Item{c}
}

func (c RelistFixedPriceItem) ParseResponse(r []byte) (Response, error) {
	return ParseXMLResponse[RelistFixedPriceItemResponse](r)
}

type RelistFixedPriceItemResponse struct {
	BaseResponse
	ItemID string `xml:"ItemID"`
}
