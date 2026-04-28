package fulfillment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type OrdersFilter struct {
	OrderIDs               []string
	CreationDateFrom       string
	CreationDateTo         string
	LastModifiedDateFrom   string
	LastModifiedDateTo     string
	OrderFulfillmentStatus string
	Limit                  int
	Offset                 int
}

func (f OrdersFilter) build() string {
	q := url.Values{}
	if len(f.OrderIDs) > 0 {
		q.Set("orderIds", strings.Join(f.OrderIDs, ","))
	}
	var filterParts []string
	if f.CreationDateFrom != "" || f.CreationDateTo != "" {
		filterParts = append(filterParts, "creationdate:["+f.CreationDateFrom+".."+f.CreationDateTo+"]")
	}
	if f.LastModifiedDateFrom != "" || f.LastModifiedDateTo != "" {
		filterParts = append(filterParts, "lastmodifieddate:["+f.LastModifiedDateFrom+".."+f.LastModifiedDateTo+"]")
	}
	if f.OrderFulfillmentStatus != "" {
		filterParts = append(filterParts, "orderfulfillmentstatus:{"+f.OrderFulfillmentStatus+"}")
	}
	if len(filterParts) > 0 {
		q.Set("filter", strings.Join(filterParts, ","))
	}
	if f.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", f.Limit))
	}
	if f.Offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", f.Offset))
	}
	if encoded := q.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

type ordersListResponse struct {
	Href   string  `json:"href"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
	Next   string  `json:"next"`
	Orders []Order `json:"orders"`
}

func (c *Client) ListOrders(ctx context.Context, filter OrdersFilter) ([]Order, error) {
	res, err := c.doer.Do(ctx, http.MethodGet, "/order"+filter.build(), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fulfillment: listOrders %d: %s", res.StatusCode, string(res.Body))
	}
	var out ordersListResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("fulfillment: decode listOrders: %w", err)
	}
	return out.Orders, nil
}

type ShippingFulfillmentRequest struct {
	LineItems       []ShippingLineItem `json:"lineItems"`
	ShippedDate     string             `json:"shippedDate,omitempty"`
	ShippingCarrier string             `json:"shippingCarrierCode,omitempty"`
	TrackingNumber  string             `json:"trackingNumber,omitempty"`
}

type ShippingLineItem struct {
	LineItemID string `json:"lineItemId"`
	Quantity   int    `json:"quantity,omitempty"`
}

func (c *Client) CreateShippingFulfillment(ctx context.Context, orderID string, req ShippingFulfillmentRequest) (string, error) {
	if orderID == "" {
		return "", fmt.Errorf("fulfillment: orderID is required")
	}
	if len(req.LineItems) == 0 {
		return "", fmt.Errorf("fulfillment: at least one line item is required")
	}
	body, _ := json.Marshal(req)
	res, err := c.doer.Do(ctx, http.MethodPost, "/order/"+url.PathEscape(orderID)+"/shipping_fulfillment", bytes.NewReader(body), "application/json")
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fulfillment: createShippingFulfillment %d: %s", res.StatusCode, string(res.Body))
	}
	if loc := res.Header.Get("Location"); loc != "" {
		if u, err := url.Parse(loc); err == nil {
			parts := strings.Split(u.Path, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1], nil
			}
		}
	}
	return "", nil
}
