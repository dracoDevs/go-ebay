package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/analytics"
	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestGetTrafficReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/traffic_report" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("date_range"); got != "[2026-01-01..2026-01-31]" {
			t.Errorf("date_range = %q", got)
		}
		if got := r.URL.Query().Get("filter"); !strings.Contains(got, "listing_ids:") {
			t.Errorf("filter = %q", got)
		}
		_, _ = w.Write([]byte(`{"records":[{"dimensionValues":[{"dimensionKey":"LISTING","value":"L1"}],"metricValues":[{"metricKey":"LISTING_VIEWS_TOTAL","value":42}]}]}`))
	}))
	defer server.Close()

	c := analytics.NewClient(auth.StaticToken("A"), analytics.WithBaseURL(server.URL))
	out, err := c.GetTrafficReport(context.Background(), analytics.TrafficReportRequest{
		ListingIDs: []string{"L1"},
		StartDate:  "2026-01-01",
		EndDate:    "2026-01-31",
	})
	if err != nil {
		t.Fatalf("GetTrafficReport: %v", err)
	}
	if len(out.Records) != 1 || out.Records[0].Metrics[0].Value != 42 {
		t.Errorf("report = %+v", out)
	}
}

func TestGetTrafficReportRequiresDates(t *testing.T) {
	c := analytics.NewClient(auth.StaticToken("A"))
	_, err := c.GetTrafficReport(context.Background(), analytics.TrafficReportRequest{})
	if err == nil || !strings.Contains(err.Error(), "StartDate and EndDate") {
		t.Errorf("expected dates-required error, got %v", err)
	}
}
