package auth

// Scope bundles for the eBay APIs this library covers. Pass these to
// RefreshTokenSource.Scopes (or ClientCredentialsSource.Scope, joined) so
// callers don't have to memorize URL strings.
//
// eBay returns a token scoped to the intersection of what the refresh token
// supports and what's requested, so over-requesting is safe.
const (
	ScopeBase                       = "https://api.ebay.com/oauth/api_scope"
	ScopeSellInventory              = "https://api.ebay.com/oauth/api_scope/sell.inventory"
	ScopeSellFulfillment            = "https://api.ebay.com/oauth/api_scope/sell.fulfillment"
	ScopeSellFulfillmentReadonly    = "https://api.ebay.com/oauth/api_scope/sell.fulfillment.readonly"
	ScopeNotificationSubscription   = "https://api.ebay.com/oauth/api_scope/commerce.notification.subscription"
	ScopeNotificationSubscriptionRO = "https://api.ebay.com/oauth/api_scope/commerce.notification.subscription.readonly"
	ScopeCommerceIdentity           = "https://api.ebay.com/oauth/api_scope/commerce.identity.readonly"
)

// Pre-built scope bundles for common use cases. Callers can compose their
// own slices if these don't match exactly.
var (
	ScopesInventory = []string{
		ScopeBase,
		ScopeSellInventory,
	}

	ScopesFulfillment = []string{
		ScopeBase,
		ScopeSellFulfillment,
		ScopeSellFulfillmentReadonly,
	}

	ScopesNotification = []string{
		ScopeBase,
		ScopeNotificationSubscription,
		ScopeNotificationSubscriptionRO,
		ScopeSellFulfillment,
		ScopeSellFulfillmentReadonly,
	}
)
