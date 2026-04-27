package trading

// ReviseFixedPriceItem updates one or more attributes of a fixed-price
// listing. Any field left at its zero value (or nil for pointer fields) is
// omitted from the request, so the caller can revise just price, just
// quantity, just SKU, etc., without inadvertently overwriting other fields.
type ReviseFixedPriceItem struct {
	ItemID                string
	SKU                   string `xml:",omitempty"`
	StartPrice            string `xml:",omitempty"`
	ConditionID           uint   `xml:",omitempty"`
	// Quantity is a pointer so callers can distinguish "don't change" (nil)
	// from "set to N" (non-nil, including 0). A bare uint with omitempty
	// would conflate Quantity=0 with "unset," and without omitempty every
	// price-only revise would unintentionally zero the quantity.
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
	ItemID string `xml:"ItemID"`
	// RelistedItemID is set when eBay issues a new listing as part of the
	// revise (rare; happens when the original listing is in a state that
	// can't be revised in place). When non-empty, ItemID points to the
	// new listing and the caller should reconcile any local references.
	RelistedItemID string `xml:"RelistedItemID,omitempty"`
}

// UintPtr is a tiny helper for the common case of passing a uint literal as
// a *uint. Useful for ReviseFixedPriceItem.Quantity.
func UintPtr(v uint) *uint { return &v }
