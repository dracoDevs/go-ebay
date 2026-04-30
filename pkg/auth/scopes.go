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
	ScopeSellMedia                  = "https://api.ebay.com/oauth/api_scope/sell.item.draft"
	ScopeSellAnalytics              = "https://api.ebay.com/oauth/api_scope/sell.analytics.readonly"
	ScopeBuyBrowse                  = "https://api.ebay.com/oauth/api_scope/buy.item"
	ScopeSellMessages               = "https://api.ebay.com/oauth/api_scope/sell.messaging"
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

	ScopesMedia = []string{
		ScopeBase,
		ScopeSellMedia,
		ScopeSellInventory,
	}

	ScopesAnalytics = []string{
		ScopeBase,
		ScopeSellAnalytics,
	}

	ScopesMessages = []string{
		ScopeBase,
		ScopeSellMessages,
	}

	ScopesBrowse = []string{
		ScopeBase,
		ScopeBuyBrowse,
	}
)
