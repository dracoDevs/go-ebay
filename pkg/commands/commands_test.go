package commands

import (
	"encoding/xml"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/ebay"
)

// TestCallNames verifies every command returns the correct eBay API call name.
func TestCallNames(t *testing.T) {
	tests := []struct {
		name     string
		command  interface{ CallName() string }
		expected string
	}{
		{"AddFixedPriceItem", AddFixedPriceItem{}, "AddFixedPriceItem"},
		{"AddItem", AddItem{}, "AddItem"},
		{"CompleteSale", CompleteSale{}, "CompleteSale"},
		{"EndItem", EndItem{}, "EndItem"},
		{"GetItem", GetItem{}, "GetItem"},
		{"GetItemTransactions", GetItemTransactions{}, "GetItemTransactions"},
		{"GetMyeBaySelling", GetMyeBaySelling{}, "GetMyeBaySelling"},
		{"GetMyMessages", GetMyMessages{}, "GetMyMessages"},
		{"GetOrders", GetOrders{}, "GetOrders"},
		{"GetTokenStatus", GetTokenStatus{}, "GetTokenStatus"},
		{"GetUser", GetUser{}, "GetUser"},
		{"ReviseFixedPriceItem", ReviseFixedPriceItem{}, "ReviseFixedPriceItem"},
		{"SetNotificationPreferences", SetNotificationPreferences{}, "SetNotificationPreferences"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.command.CallName(); got != tt.expected {
				t.Errorf("CallName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestBodyMarshalXML verifies that Body() output can be marshaled to valid XML.
func TestBodyMarshalXML(t *testing.T) {
	commands := []struct {
		name    string
		command interface{ Body() interface{} }
	}{
		{"AddFixedPriceItem", AddFixedPriceItem{Title: "Test Item", Currency: "USD"}},
		{"AddItem", AddItem{Title: "Test Item", Currency: "USD"}},
		{"CompleteSale", CompleteSale{}},
		{"EndItem", EndItem{ItemID: "123", EndingReason: NotAvailable}},
		{"GetItem", GetItem{ItemID: "123"}},
		{"GetItemTransactions", GetItemTransactions{ItemID: "123", TransactionID: "456"}},
		{"GetMyeBaySelling", GetMyeBaySelling{}},
		{"GetMyMessages", GetMyMessages{MessageIDs: MessageIDs{MessageID: "789"}, DetailLevel: "ReturnMessages"}},
		{"GetOrders", GetOrders{NumberOfDays: 30}},
		{"GetTokenStatus", GetTokenStatus{}},
		{"GetUser", GetUser{}},
		{"ReviseFixedPriceItem", ReviseFixedPriceItem{ItemID: "123", Quantity: 5}},
		{"SetNotificationPreferences", SetNotificationPreferences{}},
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

// TestParseResponseSuccess verifies that each command can parse a successful eBay response.
func TestParseResponseSuccess(t *testing.T) {
	tests := []struct {
		name    string
		command ebay.Command
		xml     string
	}{
		{
			"GetItem",
			GetItem{},
			`<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<Item><ItemID>123</ItemID><Quantity>10</Quantity></Item>
			</GetItemResponse>`,
		},
		{
			"GetUser",
			GetUser{},
			`<GetUserResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<User><UserID>testuser</UserID><EIASToken>abc123</EIASToken><Email>test@example.com</Email><Status>Confirmed</Status></User>
			</GetUserResponse>`,
		},
		{
			"GetMyMessages",
			GetMyMessages{},
			`<GetMyMessagesResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<Messages><Message><MessageID>100</MessageID><Subject>Hello</Subject><Sender>buyer1</Sender><Text>Hi there</Text></Message></Messages>
			</GetMyMessagesResponse>`,
		},
		{
			"EndItem",
			EndItem{},
			`<EndItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<EndTime>2024-01-01T00:00:00.000Z</EndTime>
			</EndItemResponse>`,
		},
		{
			"AddFixedPriceItem",
			AddFixedPriceItem{},
			`<AddFixedPriceItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<ItemID>999</ItemID>
			</AddFixedPriceItemResponse>`,
		},
		{
			"CompleteSale",
			CompleteSale{},
			`<CompleteSaleResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
			</CompleteSaleResponse>`,
		},
		{
			"GetTokenStatus",
			GetTokenStatus{},
			`<GetTokenStatusResponse xmlns="urn:ebay:apis:eBLBaseComponents">
				<Ack>Success</Ack>
				<TokenStatus><Status>Active</Status><EIASToken>token123</EIASToken></TokenStatus>
			</GetTokenStatusResponse>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.command.ParseResponse([]byte(tt.xml))
			if err != nil {
				t.Fatalf("ParseResponse() error: %v", err)
			}
			if resp.Failure() {
				t.Error("ParseResponse() returned Failure for a Success response")
			}
		})
	}
}

// TestParseResponseFailure verifies that commands correctly detect Failure ack.
func TestParseResponseFailure(t *testing.T) {
	failureXML := `<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
		<Ack>Failure</Ack>
		<Errors>
			<ShortMessage>Invalid item</ShortMessage>
			<LongMessage>The item ID is invalid.</LongMessage>
			<ErrorCode>123</ErrorCode>
		</Errors>
	</GetItemResponse>`

	cmd := GetItem{ItemID: "bad"}
	resp, err := cmd.ParseResponse([]byte(failureXML))
	if err != nil {
		t.Fatalf("ParseResponse() error: %v", err)
	}
	if !resp.Failure() {
		t.Error("expected Failure() to return true")
	}
	errs := resp.ResponseErrors()
	if len(errs) == 0 {
		t.Fatal("expected at least one error")
	}
	if errs[0].ErrorCode != 123 {
		t.Errorf("ErrorCode = %d, want 123", errs[0].ErrorCode)
	}
}

// TestParseResponseFieldValues checks that parsed fields have correct values.
func TestParseResponseFieldValues(t *testing.T) {
	t.Run("GetItem fields", func(t *testing.T) {
		xmlData := `<GetItemResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<Item>
				<ItemID>110123456789</ItemID>
				<Title>Test Product</Title>
				<Quantity>25</Quantity>
				<SellingStatus>
					<ListingStatus>Active</ListingStatus>
					<QuantitySold>5</QuantitySold>
					<CurrentPrice>29.99</CurrentPrice>
				</SellingStatus>
			</Item>
		</GetItemResponse>`

		resp, err := GetItem{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetItemResponse)
		if r.Item.ItemID != "110123456789" {
			t.Errorf("ItemID = %q, want %q", r.Item.ItemID, "110123456789")
		}
		if r.Item.Title != "Test Product" {
			t.Errorf("Title = %q, want %q", r.Item.Title, "Test Product")
		}
		if r.Item.Quantity != 25 {
			t.Errorf("Quantity = %d, want 25", r.Item.Quantity)
		}
		if r.Item.SellingStatus.CurrentPrice != 29.99 {
			t.Errorf("CurrentPrice = %f, want 29.99", r.Item.SellingStatus.CurrentPrice)
		}
	})

	t.Run("GetUser fields", func(t *testing.T) {
		xmlData := `<GetUserResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<User>
				<UserID>testuser</UserID>
				<EIASToken>nY+sHZ2PrBm</EIASToken>
				<Email>test@example.com</Email>
				<Status>Confirmed</Status>
			</User>
		</GetUserResponse>`

		resp, err := GetUser{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetUserResponse)
		if r.User.UserID != "testuser" {
			t.Errorf("UserID = %q, want %q", r.User.UserID, "testuser")
		}
		if r.User.EIASToken != "nY+sHZ2PrBm" {
			t.Errorf("EIASToken = %q, want %q", r.User.EIASToken, "nY+sHZ2PrBm")
		}
		if r.User.Email != "test@example.com" {
			t.Errorf("Email = %q, want %q", r.User.Email, "test@example.com")
		}
		if r.User.Status != "Confirmed" {
			t.Errorf("Status = %q, want %q", r.User.Status, "Confirmed")
		}
	})

	t.Run("GetMyMessages fields", func(t *testing.T) {
		xmlData := `<GetMyMessagesResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<Messages>
				<Message>
					<MessageID>100</MessageID>
					<ItemID>999</ItemID>
					<Subject>Question about item</Subject>
					<Sender>buyer1</Sender>
					<MessageType>AskSellerQuestion</MessageType>
					<Text>Is this still available?</Text>
				</Message>
			</Messages>
		</GetMyMessagesResponse>`

		resp, err := GetMyMessages{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetMyMessagesResponse)
		if len(r.Messages.Message) != 1 {
			t.Fatalf("expected 1 message, got %d", len(r.Messages.Message))
		}
		msg := r.Messages.Message[0]
		if msg.MessageID != "100" {
			t.Errorf("MessageID = %q, want %q", msg.MessageID, "100")
		}
		if msg.Subject != "Question about item" {
			t.Errorf("Subject = %q, want %q", msg.Subject, "Question about item")
		}
		if msg.Sender != "buyer1" {
			t.Errorf("Sender = %q, want %q", msg.Sender, "buyer1")
		}
		if msg.Text != "Is this still available?" {
			t.Errorf("Text = %q, want %q", msg.Text, "Is this still available?")
		}
	})

	t.Run("GetOrders fields", func(t *testing.T) {
		xmlData := `<GetOrdersResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<OrderArray>
				<Order>
					<OrderID>12345</OrderID>
					<OrderStatus>Completed</OrderStatus>
					<BuyerUserID>buyer99</BuyerUserID>
					<Total currencyID="USD">49.99</Total>
				</Order>
			</OrderArray>
		</GetOrdersResponse>`

		resp, err := GetOrders{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetOrdersResponse)
		if len(r.OrderArray.Orders) != 1 {
			t.Fatalf("expected 1 order, got %d", len(r.OrderArray.Orders))
		}
		order := r.OrderArray.Orders[0]
		if order.OrderID != "12345" {
			t.Errorf("OrderID = %q, want %q", order.OrderID, "12345")
		}
		if order.BuyerUserID != "buyer99" {
			t.Errorf("BuyerUserID = %q, want %q", order.BuyerUserID, "buyer99")
		}
		if order.Total.Value != 49.99 {
			t.Errorf("Total = %f, want 49.99", order.Total.Value)
		}
	})

	t.Run("GetMyeBaySelling fields", func(t *testing.T) {
		xmlData := `<GetMyeBaySellingResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<ActiveList>
				<ItemArray>
					<Item>
						<ItemID>555</ItemID>
						<Title>Active Listing</Title>
						<Quantity>10</Quantity>
						<QuantityAvailable>7</QuantityAvailable>
					</Item>
				</ItemArray>
				<PaginationResult>
					<TotalNumberOfPages>1</TotalNumberOfPages>
					<TotalNumberOfEntries>1</TotalNumberOfEntries>
				</PaginationResult>
			</ActiveList>
		</GetMyeBaySellingResponse>`

		resp, err := GetMyeBaySelling{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetMyeBaySellingResponse)
		if r.ActiveList == nil {
			t.Fatal("ActiveList is nil")
		}
		if r.ActiveList.ItemArray == nil {
			t.Fatal("ItemArray is nil")
		}
		if len(r.ActiveList.ItemArray.Items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(r.ActiveList.ItemArray.Items))
		}
		item := r.ActiveList.ItemArray.Items[0]
		if item.ItemID != "555" {
			t.Errorf("ItemID = %q, want %q", item.ItemID, "555")
		}
		if item.QuantityAvailable != 7 {
			t.Errorf("QuantityAvailable = %d, want 7", item.QuantityAvailable)
		}
	})

	t.Run("GetTokenStatus fields", func(t *testing.T) {
		xmlData := `<GetTokenStatusResponse xmlns="urn:ebay:apis:eBLBaseComponents">
			<Ack>Success</Ack>
			<TokenStatus>
				<Status>Active</Status>
				<EIASToken>nY+token</EIASToken>
				<ExpirationTime>2025-12-31T00:00:00.000Z</ExpirationTime>
			</TokenStatus>
		</GetTokenStatusResponse>`

		resp, err := GetTokenStatus{}.ParseResponse([]byte(xmlData))
		if err != nil {
			t.Fatalf("ParseResponse() error: %v", err)
		}
		r := resp.(GetTokenStatusResponse)
		if r.TokenStatus == nil {
			t.Fatal("TokenStatus is nil")
		}
		if r.TokenStatus.Status != "Active" {
			t.Errorf("Status = %q, want %q", r.TokenStatus.Status, "Active")
		}
		if r.TokenStatus.EIASToken != "nY+token" {
			t.Errorf("EIASToken = %q, want %q", r.TokenStatus.EIASToken, "nY+token")
		}
	})
}

// TestEndingReasonConstants verifies EndingReason typed constants.
func TestEndingReasonConstants(t *testing.T) {
	reasons := []EndingReason{
		CustomCode, Incorrect, LostOrBroken, NotAvailable,
		OtherListingError, ProductDeleted, SellToHighBidder, Sold,
	}
	for _, r := range reasons {
		if r == "" {
			t.Error("EndingReason constant is empty")
		}
	}
}

// TestBoolStrUnmarshal verifies the custom BoolStr XML unmarshaling.
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
			Value BoolStr `xml:"V"`
		}
		if err := xml.Unmarshal([]byte(`<R>`+tt.xml+`</R>`), &v); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if bool(v.Value) != tt.expected {
			t.Errorf("BoolStr(%s) = %v, want %v", tt.xml, v.Value, tt.expected)
		}
	}
}

