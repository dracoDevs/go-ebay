package ebay

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type emptyBody struct{}

// stubCommand is a minimal Command implementation for testing RunCommand.
type stubCommand struct {
	callName string
	body     interface{}
	response EbayResponse
	parseErr error
}

func (c stubCommand) CallName() string                              { return c.callName }
func (c stubCommand) Body() interface{}                             { return c.body }
func (c stubCommand) ParseResponse(r []byte) (EbayResponse, error) { return c.response, c.parseErr }

func TestRunCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "text/xml" {
			t.Errorf("Content-Type = %q, want %q", ct, "text/xml")
		}
		if cn := r.Header.Get("X-EBAY-API-CALL-NAME"); cn != "TestCall" {
			t.Errorf("X-EBAY-API-CALL-NAME = %q, want %q", cn, "TestCall")
		}
		if dev := r.Header.Get("X-EBAY-API-DEV-NAME"); dev != "dev123" {
			t.Errorf("X-EBAY-API-DEV-NAME = %q, want %q", dev, "dev123")
		}
		if app := r.Header.Get("X-EBAY-API-APP-NAME"); app != "app123" {
			t.Errorf("X-EBAY-API-APP-NAME = %q, want %q", app, "app123")
		}
		if cert := r.Header.Get("X-EBAY-API-CERT-NAME"); cert != "cert123" {
			t.Errorf("X-EBAY-API-CERT-NAME = %q, want %q", cert, "cert123")
		}
		if site := r.Header.Get("X-EBAY-API-SITEID"); site != "0" {
			t.Errorf("X-EBAY-API-SITEID = %q, want %q", site, "0")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	conf := EbayConf{
		baseUrl:   server.URL,
		DevId:     "dev123",
		AppId:     "app123",
		CertId:    "cert123",
		AuthToken: "token123",
		SiteId:    0,
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: OtherEbayResponse{Ack: "Success"},
	}

	resp, err := conf.RunCommand(cmd)
	if err != nil {
		t.Fatalf("RunCommand() error: %v", err)
	}
	if resp.Failure() {
		t.Error("expected success response")
	}
}

func TestRunCommandHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal server error")
	}))
	defer server.Close()

	conf := EbayConf{
		baseUrl:   server.URL,
		DevId:     "dev",
		AppId:     "app",
		CertId:    "cert",
		AuthToken: "token",
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: OtherEbayResponse{},
	}

	_, err := conf.RunCommand(cmd)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestRunCommandFailureAck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Failure</Ack><Errors><ShortMessage>Bad</ShortMessage><LongMessage>Bad request</LongMessage><ErrorCode>100</ErrorCode></Errors></TestCallResponse>`)
	}))
	defer server.Close()

	conf := EbayConf{
		baseUrl:   server.URL,
		DevId:     "dev",
		AppId:     "app",
		CertId:    "cert",
		AuthToken: "token",
	}

	failResp := OtherEbayResponse{
		Ack: "Failure",
		Errors: []ebayResponseError{
			{ShortMessage: "Bad", LongMessage: "Bad request", ErrorCode: 100},
		},
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: failResp,
	}

	resp, err := conf.RunCommand(cmd)
	if err == nil {
		t.Fatal("expected error for Failure ack")
	}
	if !resp.Failure() {
		t.Error("expected Failure() to return true")
	}
}

func TestSandboxAndProduction(t *testing.T) {
	conf := EbayConf{}

	sandbox := conf.Sandbox()
	if sandbox.baseUrl != "https://api.sandbox.ebay.com" {
		t.Errorf("Sandbox baseUrl = %q, want %q", sandbox.baseUrl, "https://api.sandbox.ebay.com")
	}

	prod := conf.Production()
	if prod.baseUrl != "https://api.ebay.com" {
		t.Errorf("Production baseUrl = %q, want %q", prod.baseUrl, "https://api.ebay.com")
	}
}

func TestRunCommandSendsAuthToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if len(body) == 0 {
			t.Error("empty request body")
		}
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "mySecretToken") {
			t.Error("request body does not contain auth token")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	conf := EbayConf{
		baseUrl:   server.URL,
		DevId:     "dev",
		AppId:     "app",
		CertId:    "cert",
		AuthToken: "mySecretToken",
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: OtherEbayResponse{Ack: "Success"},
	}

	_, err := conf.RunCommand(cmd)
	if err != nil {
		t.Fatalf("RunCommand() error: %v", err)
	}
}

func TestRunCommandLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	logCalls := 0
	conf := EbayConf{
		baseUrl:   server.URL,
		DevId:     "dev",
		AppId:     "app",
		CertId:    "cert",
		AuthToken: "token",
		Logger: func(args ...interface{}) {
			logCalls++
		},
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: OtherEbayResponse{Ack: "Success"},
	}

	_, err := conf.RunCommand(cmd)
	if err != nil {
		t.Fatalf("RunCommand() error: %v", err)
	}
	// Logger should be called twice: once for request, once for response
	if logCalls != 2 {
		t.Errorf("Logger called %d times, want 2", logCalls)
	}
}

func TestEbayTimestampUnmarshal(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{"RFC3339", `<T>2024-06-15T10:30:00Z</T>`, "2024-06-15"},
		{"DateTime", `<T>2024-06-15 10:30:00</T>`, "2024-06-15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v struct {
				T EbayTimestamp `xml:"T"`
			}
			if err := xml.Unmarshal([]byte(`<R>`+tt.xml+`</R>`), &v); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}
			got := v.T.Format("2006-01-02")
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("invalid format", func(t *testing.T) {
		var v struct {
			T EbayTimestamp `xml:"T"`
		}
		err := xml.Unmarshal([]byte(`<R><T>not-a-date</T></R>`), &v)
		if err == nil {
			t.Error("expected error for invalid date format")
		}
	})
}

