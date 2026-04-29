package test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/account"
	"github.com/dracoDevs/go-ebay/pkg/auth"
)

func TestGetOptedInPrograms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/program/get_opted_in_programs" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"programs":[{"programType":"SELLING_POLICY_MANAGEMENT"}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	out, err := c.GetOptedInPrograms(context.Background())
	if err != nil {
		t.Fatalf("GetOptedInPrograms: %v", err)
	}
	if len(out) != 1 || out[0].ProgramType != account.ProgramSellingPolicyManagement {
		t.Errorf("programs = %+v", out)
	}
}

func TestOptInSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/program/opt_in" {
			t.Errorf("method=%s path=%q", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got account.Program
		_ = json.Unmarshal(body, &got)
		if got.ProgramType != account.ProgramSellingPolicyManagement {
			t.Errorf("programType = %q", got.ProgramType)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	if err := c.OptIn(context.Background(), account.ProgramSellingPolicyManagement); err != nil {
		t.Fatalf("OptIn: %v", err)
	}
}

func TestOptInAlreadyOptedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":20407,"message":"already opted in"}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	if err := c.OptIn(context.Background(), account.ProgramSellingPolicyManagement); err != nil {
		t.Errorf("expected nil for already-opted-in, got %v", err)
	}
}

func TestOptInOtherErrorPropagates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12345,"message":"something else"}]}`))
	}))
	defer server.Close()

	c := account.NewClient(auth.StaticToken("A"), account.WithBaseURL(server.URL))
	err := c.OptIn(context.Background(), account.ProgramSellingPolicyManagement)
	if err == nil || !strings.Contains(err.Error(), "12345") {
		t.Errorf("expected error to surface 12345, got %v", err)
	}
}
