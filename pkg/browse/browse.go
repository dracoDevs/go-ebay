package browse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://api.ebay.com/buy/browse/v1"

type Client struct {
	doer rest.Doer
}

type Option func(*Client)

func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.doer.HTTPClient = c }
}

func WithBaseURL(u string) Option {
	return func(cl *Client) { cl.doer.BaseURL = u }
}

func NewClient(src auth.TokenSource, opts ...Option) *Client {
	c := &Client{
		doer: rest.Doer{
			TokenSource: src,
			HTTPClient:  &http.Client{Timeout: 30 * time.Second},
			BaseURL:     baseURL,
			ErrPrefix:   "browse:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Item struct {
	ItemID       string         `json:"itemId"`
	LegacyItemID string         `json:"legacyItemId,omitempty"`
	Title        string         `json:"title"`
	Price        *MoneyAmount   `json:"price,omitempty"`
	Image        *Image         `json:"image,omitempty"`
	Description  string         `json:"description,omitempty"`
	ItemWebURL   string         `json:"itemWebUrl,omitempty"`
	Categories   []Category     `json:"categories,omitempty"`
	Condition    string         `json:"condition,omitempty"`
	Brand        string         `json:"brand,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

type MoneyAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Image struct {
	ImageURL string `json:"imageUrl"`
}

type Category struct {
	CategoryID   string `json:"categoryId"`
	CategoryName string `json:"categoryName"`
}

// GetItem fetches the buyer-facing view of one listing. Use the
// "v1|<legacyItemId>|0" itemID format to look up by Trading ItemID.
func (c *Client) GetItem(ctx context.Context, itemID string) (*Item, error) {
	if itemID == "" {
		return nil, fmt.Errorf("browse: itemID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodGet, "/item/"+url.PathEscape(itemID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("browse: getItem %d: %s", res.StatusCode, string(res.Body))
	}
	var out Item
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("browse: decode getItem: %w", err)
	}
	out.Raw = res.Body
	return &out, nil
}

// GetItemByLegacyID fetches a listing by the Trading ItemID (legacy
// numeric id). Wraps the v1 lookup format.
func (c *Client) GetItemByLegacyID(ctx context.Context, legacyItemID string) (*Item, error) {
	if legacyItemID == "" {
		return nil, fmt.Errorf("browse: legacyItemID is required")
	}
	return c.GetItem(ctx, "v1|"+legacyItemID+"|0")
}
