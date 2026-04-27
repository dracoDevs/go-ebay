package test

import (
	"encoding/xml"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/trading"
)

func TestCallNames(t *testing.T) {
	tests := []struct {
		name     string
		command  interface{ CallName() string }
		expected string
	}{
		{"AddFixedPriceItem", trading.AddFixedPriceItem{}, "AddFixedPriceItem"},
		{"AddItem", trading.AddItem{}, "AddItem"},
		{"CompleteSale", trading.CompleteSale{}, "CompleteSale"},
		{"EndItem", trading.EndItem{}, "EndItem"},
		{"GetItem", trading.GetItem{}, "GetItem"},
		{"GetItemTransactions", trading.GetItemTransactions{}, "GetItemTransactions"},
		{"GetMyeBaySelling", trading.GetMyeBaySelling{}, "GetMyeBaySelling"},
		{"GetMyMessages", trading.GetMyMessages{}, "GetMyMessages"},
		{"GetOrders", trading.GetOrders{}, "GetOrders"},
		{"GetTokenStatus", trading.GetTokenStatus{}, "GetTokenStatus"},
		{"GetUser", trading.GetUser{}, "GetUser"},
		{"ReviseFixedPriceItem", trading.ReviseFixedPriceItem{}, "ReviseFixedPriceItem"},
		{"SetNotificationPreferences", trading.SetNotificationPreferences{}, "SetNotificationPreferences"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.command.CallName(); got != tt.expected {
				t.Errorf("CallName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBodyMarshalXML(t *testing.T) {
	commands := []struct {
		name    string
		command interface{ Body() interface{} }
	}{
		{"AddFixedPriceItem", trading.AddFixedPriceItem{Title: "Test Item", Currency: "USD"}},
		{"AddItem", trading.AddItem{Title: "Test Item", Currency: "USD"}},
		{"CompleteSale", trading.CompleteSale{}},
		{"EndItem", trading.EndItem{ItemID: "123", EndingReason: trading.NotAvailable}},
		{"GetItem", trading.GetItem{ItemID: "123"}},
		{"GetItemTransactions", trading.GetItemTransactions{ItemID: "123", TransactionID: "456"}},
		{"GetMyeBaySelling", trading.GetMyeBaySelling{}},
		{"GetMyMessages", trading.GetMyMessages{MessageIDs: trading.MessageIDs{MessageID: "789"}, DetailLevel: "ReturnMessages"}},
		{"GetOrders", trading.GetOrders{NumberOfDays: 30}},
		{"GetTokenStatus", trading.GetTokenStatus{}},
		{"GetUser", trading.GetUser{}},
		{"ReviseFixedPriceItem", trading.ReviseFixedPriceItem{ItemID: "123", Quantity: trading.UintPtr(5)}},
		{"SetNotificationPreferences", trading.SetNotificationPreferences{}},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			body := tt.command.Body()
			if _, err := xml.Marshal(body); err != nil {
				t.Errorf("Body() cannot be marshaled to XML: %v", err)
			}
		})
	}
}

func TestParseResponseSuccess(t *testing.T) {
	tests := []struct {
		name    string
		command trading.Command
		xml     string
	}{
		{"GetItem", trading.GetItem{}, `<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><Item><ItemID>123</ItemID><Quantity>10</Quantity></Item></GetItemResponse>`},
		{"GetUser", trading.GetUser{}, `<GetUserResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><User><UserID>testuser</UserID><EIASToken>abc123</EIASToken><Email>test@example.com</Email><Status>Confirmed</Status></User></GetUserResponse>`},
		{"GetMyMessages", trading.GetMyMessages{}, `<GetMyMessagesResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><Messages><Message><MessageID>100</MessageID><Subject>Hello</Subject><Sender>buyer1</Sender><Text>Hi there</Text></Message></Messages></GetMyMessagesResponse>`},
		{"EndItem", trading.EndItem{}, `<EndItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><EndTime>2024-01-01T00:00:00.000Z</EndTime></EndItemResponse>`},
		{"AddFixedPriceItem", trading.AddFixedPriceItem{}, `<AddFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><ItemID>999</ItemID></AddFixedPriceItemResponse>`},
		{"CompleteSale", trading.CompleteSale{}, `<CompleteSaleResponse xmlns="urn:ebay:apis:eBLBaseComponents"><Ack>Success</Ack></CompleteSaleResponse>`},
		{"GetTokenStatus", trading.GetTokenStatus{}, `<GetTokenStatusResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack><TokenStatus><Status>Active</Status><EIASToken>token123</EIASToken></TokenStatus></GetTokenStatusResponse>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.command.ParseResponse([]byte(tt.xml))
			if err != nil {
				t.Fatalf("ParseResponse: %v", err)
			}
			if resp.Failure() {
				t.Error("ParseResponse returned Failure for a Success response")
			}
		})
	}
}

func TestParseResponseFailure(t *testing.T) {
	failureXML := `<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
		<Ack>Failure</Ack>
		<Errors><ShortMessage>Invalid item</ShortMessage><LongMessage>The item ID is invalid.</LongMessage><ErrorCode>123</ErrorCode></Errors>
	</GetItemResponse>`

	resp, err := trading.GetItem{ItemID: "bad"}.ParseResponse([]byte(failureXML))
	if err != nil {
		t.Fatalf("ParseResponse: %v", err)
	}
	if !resp.Failure() {
		t.Error("expected Failure() == true")
	}
	errs := resp.ResponseErrors()
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if errs[0].ErrorCode != 123 {
		t.Errorf("ErrorCode = %d, want 123", errs[0].ErrorCode)
	}
}

func TestParseResponseFieldValues(t *testing.T) {
	t.Run("GetItem fields", func(t *testing.T) {
		xmlData := `<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<Item><ItemID>110123456789</ItemID><Title>Test Product</Title><Quantity>25</Quantity>
				<SellingStatus><ListingStatus>Active</ListingStatus><QuantitySold>5</QuantitySold><CurrentPrice>29.99</CurrentPrice></SellingStatus>
			</Item></GetItemResponse>`

		resp, err := trading.GetItem{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetItemResponse)
		if r.Item.ItemID != "110123456789" {
			t.Errorf("ItemID = %q", r.Item.ItemID)
		}
		if r.Item.Title != "Test Product" {
			t.Errorf("Title = %q", r.Item.Title)
		}
		if r.Item.Quantity != 25 {
			t.Errorf("Quantity = %d", r.Item.Quantity)
		}
		if r.Item.SellingStatus.CurrentPrice != 29.99 {
			t.Errorf("CurrentPrice = %f", r.Item.SellingStatus.CurrentPrice)
		}
	})

	t.Run("GetUser fields", func(t *testing.T) {
		xmlData := `<GetUserResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<User><UserID>testuser</UserID><EIASToken>nY+sHZ2PrBm</EIASToken><Email>test@example.com</Email><Status>Confirmed</Status></User>
		</GetUserResponse>`
		resp, err := trading.GetUser{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetUserResponse)
		if r.User.UserID != "testuser" {
			t.Errorf("UserID = %q", r.User.UserID)
		}
		if r.User.EIASToken != "nY+sHZ2PrBm" {
			t.Errorf("EIASToken = %q", r.User.EIASToken)
		}
	})

	t.Run("GetMyMessages fields", func(t *testing.T) {
		xmlData := `<GetMyMessagesResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<Messages><Message><MessageID>100</MessageID><ItemID>999</ItemID><Subject>Question about item</Subject><Sender>buyer1</Sender><MessageType>AskSellerQuestion</MessageType><Text>Is this still available?</Text></Message></Messages>
		</GetMyMessagesResponse>`
		resp, err := trading.GetMyMessages{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetMyMessagesResponse)
		if len(r.Messages.Message) != 1 {
			t.Fatalf("expected 1 message, got %d", len(r.Messages.Message))
		}
		msg := r.Messages.Message[0]
		if msg.MessageID != "100" || msg.Subject != "Question about item" || msg.Sender != "buyer1" {
			t.Errorf("msg = %+v", msg)
		}
	})

	t.Run("GetOrders fields", func(t *testing.T) {
		xmlData := `<GetOrdersResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<OrderArray><Order><OrderID>12345</OrderID><OrderStatus>Completed</OrderStatus><BuyerUserID>buyer99</BuyerUserID><Total currencyID="USD">49.99</Total></Order></OrderArray>
		</GetOrdersResponse>`
		resp, err := trading.GetOrders{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetOrdersResponse)
		if len(r.OrderArray.Orders) != 1 {
			t.Fatalf("expected 1 order, got %d", len(r.OrderArray.Orders))
		}
		order := r.OrderArray.Orders[0]
		if order.OrderID != "12345" || order.BuyerUserID != "buyer99" || order.Total.Value != 49.99 {
			t.Errorf("order = %+v", order)
		}
	})

	t.Run("GetMyeBaySelling fields", func(t *testing.T) {
		xmlData := `<GetMyeBaySellingResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<ActiveList><ItemArray><Item><ItemID>555</ItemID><Title>Active Listing</Title><Quantity>10</Quantity><QuantityAvailable>7</QuantityAvailable></Item></ItemArray>
				<PaginationResult><TotalNumberOfPages>1</TotalNumberOfPages><TotalNumberOfEntries>1</TotalNumberOfEntries></PaginationResult>
			</ActiveList></GetMyeBaySellingResponse>`
		resp, err := trading.GetMyeBaySelling{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetMyeBaySellingResponse)
		if r.ActiveList == nil || r.ActiveList.ItemArray == nil {
			t.Fatal("ActiveList/ItemArray is nil")
		}
		if len(r.ActiveList.ItemArray.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(r.ActiveList.ItemArray.Items))
		}
		item := r.ActiveList.ItemArray.Items[0]
		if item.ItemID != "555" || item.QuantityAvailable != 7 {
			t.Errorf("item = %+v", item)
		}
	})

	t.Run("GetTokenStatus fields", func(t *testing.T) {
		xmlData := `<GetTokenStatusResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<TokenStatus><Status>Active</Status><EIASToken>nY+token</EIASToken><ExpirationTime>2025-12-31T00:00:00.000Z</ExpirationTime></TokenStatus>
		</GetTokenStatusResponse>`
		resp, err := trading.GetTokenStatus{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse: %v", err)
		}
		r := resp.(trading.GetTokenStatusResponse)
		if r.TokenStatus == nil {
			t.Fatal("TokenStatus is nil")
		}
		if r.TokenStatus.Status != "Active" || r.TokenStatus.EIASToken != "nY+token" {
			t.Errorf("TokenStatus = %+v", r.TokenStatus)
		}
	})
}

func TestEndingReasonConstants(t *testing.T) {
	reasons := []trading.EndingReason{
		trading.CustomCode, trading.Incorrect, trading.LostOrBroken, trading.NotAvailable,
		trading.OtherListingError, trading.ProductDeleted, trading.SellToHighBidder, trading.Sold,
	}
	for _, r := range reasons {
		if r == "" {
			t.Error("EndingReason constant is empty")
		}
	}
}

func TestBoolStrUnmarshal(t *testing.T) {
	tests := []struct {
		xml      string
		expected bool
	}{
		{`<V>true</V>`, true},
		{`<V>false</V>`, false},
		{`<V>anything</V>`, false},
	}

	for _, tt := range tests {
		var v struct {
			Value trading.BoolStr `xml:"V"`
		}
		if err := xml.Unmarshal([]byte(`<R>`+tt.xml+`</R>`), &v); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if bool(v.Value) != tt.expected {
			t.Errorf("BoolStr(%s) = %v, want %v", tt.xml, v.Value, tt.expected)
		}
	}
}
