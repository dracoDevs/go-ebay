package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/internal/rest"
)

const baseURL = "https://api.ebay.com/sell/messages/v1"

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
			ErrPrefix:   "messages:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Conversation struct {
	ConversationID string    `json:"conversationId"`
	BuyerUsername  string    `json:"buyerUsername"`
	SellerUsername string    `json:"sellerUsername"`
	Subject        string    `json:"subject,omitempty"`
	Status         string    `json:"status,omitempty"`
	LastMessageAt  string    `json:"lastMessageAt,omitempty"`
	Messages       []Message `json:"messages,omitempty"`
}

type Message struct {
	MessageID  string `json:"messageId"`
	From       string `json:"from,omitempty"`
	To         string `json:"to,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Body       string `json:"body,omitempty"`
	SentAt     string `json:"sentAt,omitempty"`
	ItemID     string `json:"itemId,omitempty"`
}

func (c *Client) GetConversation(ctx context.Context, conversationID string) (*Conversation, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("messages: conversationID is required")
	}
	res, err := c.doer.Do(ctx, http.MethodGet, "/conversation/"+url.PathEscape(conversationID), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("messages: getConversation %d: %s", res.StatusCode, string(res.Body))
	}
	var out Conversation
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("messages: decode getConversation: %w", err)
	}
	return &out, nil
}

type SendMessageRequest struct {
	Body    string `json:"body"`
	Subject string `json:"subject,omitempty"`
	ItemID  string `json:"itemId,omitempty"`
}

// SendMessage posts a message in an existing conversation. eBay's Sell
// Messages API requires a conversationId; for buyer-on-itemId messaging
// the caller must first resolve the conversation (typically by replying
// in-thread to an existing buyer message). For unsolicited seller-to-
// buyer messages on a listing context, callers should fall back to the
// Trading SendMessageToBuyer path until eBay provides a REST equivalent
// for that flow.
func (c *Client) SendMessage(ctx context.Context, conversationID string, req SendMessageRequest) error {
	if conversationID == "" {
		return fmt.Errorf("messages: conversationID is required")
	}
	if req.Body == "" {
		return fmt.Errorf("messages: body is required")
	}
	body, _ := json.Marshal(req)
	res, err := c.doer.Do(ctx, http.MethodPost, "/conversation/"+url.PathEscape(conversationID)+"/send_message", bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("messages: sendMessage %d: %s", res.StatusCode, string(res.Body))
	}
	return nil
}

// IsNotEnrolled reports whether err looks like a "Sell Messages API not
// available for this seller / region" failure (HTTP 403/404). Callers
// can use this to fall back to the Trading messaging path while the
// Sell Messages API rolls out region-by-region.
func IsNotEnrolled(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, " 403:") || strings.Contains(msg, " 404:") || strings.Contains(msg, "not enrolled")
}
