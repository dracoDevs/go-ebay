package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/auth"
	"github.com/dracoDevs/go-ebay/pkg/media"
)

func TestCreateImageFromURLBodyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/image/create_from_url" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"imageUrl":"https://src.example/x.jpg"`) {
			t.Errorf("body = %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"imageId":"img-1","imageUrl":"https://i.ebayimg.com/images/g/abc/s-l1600.jpg"}`))
	}))
	defer server.Close()

	c := media.NewClient(auth.StaticToken("A"), media.WithBaseURL(server.URL))
	got, err := c.CreateImageFromURL(context.Background(), "https://src.example/x.jpg")
	if err != nil {
		t.Fatalf("CreateImageFromURL: %v", err)
	}
	if got != "https://i.ebayimg.com/images/g/abc/s-l1600.jpg" {
		t.Errorf("got %q", got)
	}
}

func TestCreateImageFromURLEmptyInput(t *testing.T) {
	c := media.NewClient(auth.StaticToken("A"))
	_, err := c.CreateImageFromURL(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "imageURL is required") {
		t.Errorf("expected required error, got %v", err)
	}
}

func TestCreateImageFromURLPropagatesNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":2003}]}`))
	}))
	defer server.Close()

	c := media.NewClient(auth.StaticToken("A"), media.WithBaseURL(server.URL))
	_, err := c.CreateImageFromURL(context.Background(), "https://src.example/x.jpg")
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 error, got %v", err)
	}
}
