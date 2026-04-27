package trading

type ReviseFixedPriceItem struct {
	ItemID      string
	SKU         string `xml:",omitempty"`
	StartPrice  string `xml:",omitempty"`
	ConditionID uint   `xml:",omitempty"`
	// nil = don't change. Bare uint without omitempty would unintentionally
	// zero quantity on every price-only revise.
	Quantity              *uint
	Title                 string           `xml:",omitempty"`
	Description           string           `xml:",omitempty"`
	PayPalEmailAddress    string           `xml:",omitempty"`
	PictureDetails        *PictureDetails  `xml:",omitempty"`
	ShippingDetails       *ShippingDetails `xml:",omitempty"`
	PrimaryCategory       *PrimaryCategory
	ReturnPolicy          *ReturnPolicy          `xml:",omitempty"`
	ProductListingDetails *ProductListingDetails `xml:",omitempty"`
	ItemSpecifics         *ItemSpecifics         `xml:",omitempty"`
}

func (c ReviseFixedPriceItem) CallName() string { return "ReviseFixedPriceItem" }

func (c ReviseFixedPriceItem) Body() interface{} {
	type Item struct{ ReviseFixedPriceItem }
	return Item{c}
}

func (c ReviseFixedPriceItem) ParseResponse(r []byte) (Response, error) {
	return ParseXMLResponse[ReviseFixedPriceItemResponse](r)
}

type ReviseFixedPriceItemResponse struct {
	BaseResponse
	ItemID         string `xml:"ItemID"`
	RelistedItemID string `xml:"RelistedItemID,omitempty"`
}

func UintPtr(v uint) *uint { return &v }
