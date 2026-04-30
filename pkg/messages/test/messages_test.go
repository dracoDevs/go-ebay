package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/messages"
)

func TestGetConversation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/conversation/CONV-1" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"conversationId":"CONV-1","buyerUsername":"buyer1","messages":[{"messageId":"M1","body":"hello"}]}`))
	}))
	defer server.Close()

	c := messages.NewClient(auth.StaticToken("A"), messages.WithBaseURL(server.URL))
	out, err := c.GetConversation(context.Background(), "CONV-1")
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if out.ConversationID != "CONV-1" || out.BuyerUsername != "buyer1" {
		t.Errorf("conversation = %+v", out)
	}
	if len(out.Messages) != 1 || out.Messages[0].MessageID != "M1" {
		t.Errorf("messages = %+v", out.Messages)
	}
}

func TestSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/conversation/CONV-1/send_message" || r.Method != http.MethodPost {
			t.Errorf("method=%s path=%q", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := messages.NewClient(auth.StaticToken("A"), messages.WithBaseURL(server.URL))
	err := c.SendMessage(context.Background(), "CONV-1", messages.SendMessageRequest{Body: "thanks"})
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
}

func TestSendMessageRequiresBody(t *testing.T) {
	c := messages.NewClient(auth.StaticToken("A"))
	err := c.SendMessage(context.Background(), "CONV", messages.SendMessageRequest{})
	if err == nil || !strings.Contains(err.Error(), "body is required") {
		t.Errorf("expected body-required error, got %v", err)
	}
}

func TestIsNotEnrolled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1500,"message":"not enrolled"}]}`))
	}))
	defer server.Close()

	c := messages.NewClient(auth.StaticToken("A"), messages.WithBaseURL(server.URL))
	_, err := c.GetConversation(context.Background(), "X")
	if err == nil {
		t.Fatal("expected error")
	}
	if !messages.IsNotEnrolled(err) {
		t.Errorf("IsNotEnrolled(%v) = false, want true", err)
	}
}
