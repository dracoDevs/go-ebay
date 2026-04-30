package analytics

import (
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

const baseURL = "https://api.ebay.com/sell/analytics/v1"

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
			ErrPrefix:   "analytics:",
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

type TrafficReportRequest struct {
	ListingIDs []string
	StartDate  string
	EndDate    string
	Marketplace string
	Metrics    []string
}

type TrafficReport struct {
	Records           []TrafficRecord `json:"records"`
	StartDate         string          `json:"startDate,omitempty"`
	EndDate           string          `json:"endDate,omitempty"`
	LastUpdatedDate   string          `json:"lastUpdatedDate,omitempty"`
}

type TrafficRecord struct {
	Dimensions []TrafficDimension `json:"dimensionValues"`
	Metrics    []TrafficMetric    `json:"metricValues"`
}

type TrafficDimension struct {
	Key   string `json:"dimensionKey"`
	Value string `json:"value"`
}

type TrafficMetric struct {
	Key   string  `json:"metricKey"`
	Value float64 `json:"value"`
}

func (c *Client) GetTrafficReport(ctx context.Context, req TrafficReportRequest) (*TrafficReport, error) {
	if req.StartDate == "" || req.EndDate == "" {
		return nil, fmt.Errorf("analytics: StartDate and EndDate are required")
	}
	q := url.Values{}
	q.Set("dimension", "LISTING")
	if req.StartDate != "" {
		q.Set("date_range", fmt.Sprintf("[%s..%s]", req.StartDate, req.EndDate))
	}
	if req.Marketplace != "" {
		q.Set("marketplace_ids", "{"+req.Marketplace+"}")
	} else {
		q.Set("marketplace_ids", "{EBAY_US}")
	}
	if len(req.ListingIDs) > 0 {
		q.Set("filter", "listing_ids:{"+strings.Join(req.ListingIDs, "|")+"}")
	}
	if len(req.Metrics) > 0 {
		q.Set("metric", strings.Join(req.Metrics, ","))
	} else {
		q.Set("metric", "CLICK_THROUGH_RATE,LISTING_IMPRESSION_TOTAL,LISTING_VIEWS_TOTAL,SALES_CONVERSION_RATE")
	}
	res, err := c.doer.Do(ctx, http.MethodGet, "/traffic_report?"+q.Encode(), nil, "")
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("analytics: getTrafficReport %d: %s", res.StatusCode, string(res.Body))
	}
	var out TrafficReport
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, fmt.Errorf("analytics: decode getTrafficReport: %w", err)
	}
	return &out, nil
}
