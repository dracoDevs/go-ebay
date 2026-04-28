package auth

const (
	ScopeBase                       = "https://api.ebay.com/oauth/api_scope"
	ScopeSellInventory              = "https://api.ebay.com/oauth/api_scope/sell.inventory"
	ScopeSellInventoryReadonly      = "https://api.ebay.com/oauth/api_scope/sell.inventory.readonly"
	ScopeSellAccount                = "https://api.ebay.com/oauth/api_scope/sell.account"
	ScopeSellAccountReadonly        = "https://api.ebay.com/oauth/api_scope/sell.account.readonly"
	ScopeSellFulfillment            = "https://api.ebay.com/oauth/api_scope/sell.fulfillment"
	ScopeSellFulfillmentReadonly    = "https://api.ebay.com/oauth/api_scope/sell.fulfillment.readonly"
	ScopeNotificationSubscription   = "https://api.ebay.com/oauth/api_scope/commerce.notification.subscription"
	ScopeNotificationSubscriptionRO = "https://api.ebay.com/oauth/api_scope/commerce.notification.subscription.readonly"
	ScopeCommerceIdentity           = "https://api.ebay.com/oauth/api_scope/commerce.identity.readonly"
)

var (
	ScopesInventory = []string{
		ScopeBase,
		ScopeSellInventory,
	}

	ScopesAccount = []string{
		ScopeBase,
		ScopeSellAccount,
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

	ScopesIdentity = []string{
		ScopeBase,
		ScopeCommerceIdentity,
	}

	ScopesListingCreate = []string{
		ScopeBase,
		ScopeSellInventory,
		ScopeSellAccount,
	}
)
