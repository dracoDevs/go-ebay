package notification

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestListTopicsFollowsPagination(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch calls {
		case 1:
			if r.URL.Path != "/topic" {
				t.Errorf("first call path = %q", r.URL.Path)
			}
			next := "http://" + r.Host + "/topic?cursor=2"
			fmt := `{"topics":[{"topicId":"T1","status":"ENABLED"}],"next":%q}`
			_, _ = w.Write([]byte(jsonF(fmt, next)))
		case 2:
			if r.URL.Query().Get("cursor") != "2" {
				t.Errorf("second call cursor = %q", r.URL.Query().Get("cursor"))
			}
			_, _ = w.Write([]byte(`{"topics":[{"topicId":"T2","status":"ENABLED"}]}`))
		default:
			t.Fatalf("unexpected call %d", calls)
		}
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	topics, err := c.ListTopics(context.Background())
	if err != nil {
		t.Fatalf("ListTopics: %v", err)
	}
	if len(topics) != 2 || topics[0].TopicID != "T1" || topics[1].TopicID != "T2" {
		t.Errorf("topics = %+v", topics)
	}
}

func TestListDestinations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/destination" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"destinations":[{"destinationId":"D1","status":"ENABLED","deliveryConfig":{"endpoint":"https://x.example.com/hook","verificationToken":"V"}}]}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	dests, err := c.ListDestinations(context.Background())
	if err != nil {
		t.Fatalf("ListDestinations: %v", err)
	}
	if len(dests) != 1 || dests[0].DestinationID != "D1" || dests[0].DeliveryConfig.Endpoint != "https://x.example.com/hook" {
		t.Errorf("dests = %+v", dests)
	}
}

func TestCreateDestinationParsesLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req CreateDestinationRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode req: %v", err)
		}
		if req.Name != "primary" || req.DeliveryConfig.Endpoint != "https://x.example.com/hook" {
			t.Errorf("req = %+v", req)
		}
		w.Header().Set("Location", "/commerce/notification/v1/destination/D-NEW-123")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	id, err := c.CreateDestination(context.Background(), CreateDestinationRequest{
		Name: "primary", Status: "ENABLED",
		DeliveryConfig: DeliveryConfig{Endpoint: "https://x.example.com/hook", VerificationToken: "V"},
	})
	if err != nil {
		t.Fatalf("CreateDestination: %v", err)
	}
	if id != "D-NEW-123" {
		t.Errorf("id = %q", id)
	}
}

func TestUpdateConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/config" {
			t.Errorf("method=%s path=%q", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"alertEmail":"ops@x.com"`) {
			t.Errorf("body = %s", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))
	if err := c.UpdateConfig(context.Background(), "ops@x.com"); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}
}

func TestListAndCreateSubscription(t *testing.T) {
	subs := []Subscription{{SubscriptionID: "S1", TopicID: "ORDER_CONFIRMATION", Status: "ENABLED"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/subscription" && r.Method == http.MethodGet {
			body, _ := json.Marshal(subscriptionListResponse{Subscriptions: subs})
			_, _ = w.Write(body)
			return
		}
		if r.URL.Path == "/subscription" && r.Method == http.MethodPost {
			w.Header().Set("Location", "/commerce/notification/v1/subscription/S-NEW")
			w.WriteHeader(http.StatusCreated)
			return
		}
		t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("ACCESS"), WithBaseURL(server.URL))

	got, err := c.ListSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("ListSubscriptions: %v", err)
	}
	if len(got) != 1 || got[0].SubscriptionID != "S1" {
		t.Errorf("subs = %+v", got)
	}

	id, err := c.CreateSubscription(context.Background(), CreateSubscriptionRequest{
		TopicID: "ORDER_CONFIRMATION", Status: "ENABLED", DestinationID: "D",
		Payload: Payload{Format: "JSON", SchemaVersion: "1.0", DeliveryProtocol: "HTTPS"},
	})
	if err != nil {
		t.Fatalf("CreateSubscription: %v", err)
	}
	if id != "S-NEW" {
		t.Errorf("id = %q", id)
	}
}

func TestEnableDisableTestSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/enable"):
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/disable"):
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/test"):
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"notificationId":"N-1"}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("A"), WithBaseURL(server.URL))
	if err := c.EnableSubscription(context.Background(), "S1"); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	if err := c.DisableSubscription(context.Background(), "S1"); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	id, err := c.TestSubscription(context.Background(), "S1")
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if id != "N-1" {
		t.Errorf("notificationId = %q", id)
	}
}

func TestPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1001}]}`))
	}))
	defer server.Close()

	c := NewClient(auth.StaticToken("A"), WithBaseURL(server.URL))
	_, err := c.ListSubscriptions(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401, got %v", err)
	}
}

// jsonF lets the test inline a single %q replacement without bringing in fmt.
func jsonF(fmtStr, arg string) string {
	q := strings.ReplaceAll(arg, `"`, `\"`)
	return strings.Replace(fmtStr, "%q", `"`+q+`"`, 1)
}
