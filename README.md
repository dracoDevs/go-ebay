# go-ebay

Go client library for eBay's seller-side APIs. Covers the four surfaces we
hit in production:

| Package | API | Wire format | Auth |
|---|---|---|---|
| [`pkg/trading`](pkg/trading) | Trading API | XML over `/ws/api.dll` | `eBayAuthToken` (long-lived) |
| [`pkg/inventory`](pkg/inventory) | Sell Inventory v1 | JSON | OAuth user token |
| [`pkg/fulfillment`](pkg/fulfillment) | Sell Fulfillment v1 | JSON | OAuth user token |
| [`pkg/notification`](pkg/notification) | Commerce Notification v1 | JSON | OAuth (app for topics/destinations, user for subscriptions) |
| [`pkg/auth`](pkg/auth) | OAuth identity | form-encoded | n/a (mints tokens) |

## Install

```bash
go get github.com/dracoDevs/go-ebay
```

## Quick start

### Trading (XML)

```go
import "github.com/dracoDevs/go-ebay/pkg/trading"

conf := trading.Conf{
    DevId:     os.Getenv("EBAY_DEV_ID"),
    AppId:     os.Getenv("EBAY_API_KEY"),
    CertId:    os.Getenv("EBAY_CERT_ID"),
    AuthToken: user.AuthToken,
}.Production()

resp, err := conf.RunCommand(trading.ReviseFixedPriceItem{
    ItemID:   "123456789012",
    Quantity: trading.UintPtr(5),
})
if err != nil { /* ... */ }
revised := resp.(trading.ReviseFixedPriceItemResponse)
fmt.Println(revised.Ack, revised.RelistedItemID)
```

The Trading client follows a `Command` interface: every operation is a
struct that knows how to marshal itself, name itself, and parse its
response. Run any command via `Conf.RunCommand`. See
[`pkg/trading`](pkg/trading) for the full list (`AddFixedPriceItem`,
`GetItem`, `GetOrders`, `EndItem`, `RelistFixedPriceItem`,
`SetNotificationPreferences`, etc.).

### Inventory (REST)

```go
import (
    "github.com/dracoDevs/go-ebay/pkg/auth"
    "github.com/dracoDevs/go-ebay/pkg/inventory"
)

src := &auth.RefreshTokenSource{
    AppID:        os.Getenv("EBAY_API_KEY"),
    CertID:       os.Getenv("EBAY_CERT_ID"),
    RefreshToken: user.RefreshToken,
    Scopes:       auth.ScopesInventory,
}
client := inventory.NewClient(src)

results, err := client.BulkMigrateListings(ctx, []string{"L1", "L2"})
// results[i].InventoryItems[0].SKU / .OfferID

qty := 5
price := 9.99
_, err = client.BulkUpdatePriceQuantity(ctx, []inventory.PriceQuantityUpdate{{
    OfferID:  "OFFER1",
    SKU:      "SKU1",
    Quantity: &qty,
    Price:    &price,
}})
```

`BulkMigrateListings` accepts up to 5 listing IDs per call. eBay can return
207 (Multi-Status) when some succeed and some fail; the per-listing
breakdown comes back in the response slice and only transport-level
failures surface as a Go error.

### Fulfillment (REST)

```go
import (
    "github.com/dracoDevs/go-ebay/pkg/auth"
    "github.com/dracoDevs/go-ebay/pkg/fulfillment"
)

client := fulfillment.NewClient(&auth.RefreshTokenSource{
    AppID:        os.Getenv("EBAY_API_KEY"),
    CertID:       os.Getenv("EBAY_CERT_ID"),
    RefreshToken: user.RefreshToken,
    Scopes:       auth.ScopesFulfillment,
})

order, err := client.GetOrder(ctx, "12-34567-89012")
fmt.Println(order.OrderID, order.PaymentSummary.TotalDueSeller.FloatValue())
```

Use `WithMarketplace("EBAY_GB")` (etc.) to target a non-US site; default is
`EBAY_US`.

### Notification (REST)

App-level (topics, destinations, account config):

```go
import (
    "github.com/dracoDevs/go-ebay/pkg/auth"
    "github.com/dracoDevs/go-ebay/pkg/notification"
)

app := notification.NewClient(&auth.ClientCredentialsSource{
    AppID:  os.Getenv("EBAY_API_KEY"),
    CertID: os.Getenv("EBAY_CERT_ID"),
    Scope:  auth.ScopeBase,
})

topics, _ := app.ListTopics(ctx)
destID, _ := app.CreateDestination(ctx, notification.CreateDestinationRequest{
    Name: "primary", Status: "ENABLED",
    DeliveryConfig: notification.DeliveryConfig{
        Endpoint:          "https://api.example.com/ebay/webhook",
        VerificationToken: os.Getenv("EBAY_VERIFICATION_TOKEN"),
    },
})
```

User-level (subscriptions):

```go
user := notification.NewClient(&auth.RefreshTokenSource{
    AppID:        os.Getenv("EBAY_API_KEY"),
    CertID:       os.Getenv("EBAY_CERT_ID"),
    RefreshToken: someUser.RefreshToken,
    Scopes:       auth.ScopesNotification,
})

subs, _ := user.ListSubscriptions(ctx)
subID, _ := user.CreateSubscription(ctx, notification.CreateSubscriptionRequest{
    TopicID:       "ORDER_CONFIRMATION",
    Status:        "ENABLED",
    DestinationID: destID,
    Payload:       notification.Payload{Format: "JSON", SchemaVersion: "1.0", DeliveryProtocol: "HTTPS"},
})
_, _ = user.TestSubscription(ctx, subID)
```

The same `Client` type handles both flows; the only difference is which
`auth.TokenSource` you hand it.

## Auth (`pkg/auth`)

Every REST package takes an `auth.TokenSource` so the call site doesn't
care whether the token is static, refresh-derived, or app-level. The two
production sources cache mint results and only re-mint when the cached
token is within ~60s of expiring, so tight loops don't hammer the eBay
identity endpoint.

| Source | Purpose |
|---|---|
| `auth.StaticToken("...")` | Tests, CLIs, anywhere you already have a bearer string. |
| `auth.RefreshTokenSource{...}` | A specific seller's user token. Caches with expiry. |
| `auth.ClientCredentialsSource{...}` | Application-level token. Caches with expiry. |

Pre-built scope bundles:

- `auth.ScopesInventory`
- `auth.ScopesFulfillment`
- `auth.ScopesNotification`

Plus individual `Scope*` constants if you need to compose your own slice.

## Customizing the HTTP client

Every REST `NewClient` accepts `WithHTTPClient(*http.Client)` and
`WithBaseURL(string)` options. Use `WithHTTPClient` to share a transport
across the library and your own code (custom dialer, proxy, retry
middleware). Use `WithBaseURL` for sandbox or `httptest` servers.

```go
inv := inventory.NewClient(src,
    inventory.WithHTTPClient(myClient),
    inventory.WithBaseURL("https://api.sandbox.ebay.com/sell/inventory/v1"),
)
```

The Trading client (`trading.Conf`) has built-in `Sandbox()` / `Production()`
toggles instead, since its base URL is more stable across the API surface.

## Tests

Each package's tests live alongside it (Go convention):

```
pkg/auth/auth_test.go
pkg/inventory/inventory_test.go
pkg/fulfillment/fulfillment_test.go
pkg/notification/notification_test.go
pkg/trading/client_test.go
pkg/trading/commands_test.go
pkg/trading/revise_fixed_price_item_test.go
```

The REST packages use `httptest.NewServer` and `WithBaseURL` to verify
request shape and response decoding without hitting eBay. Run them with:

```bash
go test ./...
```

## Compatibility

This library is at v0; the public surface may still change. Tag `v0.0.1`
captures the pre-rename layout (`pkg/ebay` + `pkg/commands`) for anyone who
needs to pin while migrating.

## License

MIT.
