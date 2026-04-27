package test

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracoDevs/go-ebay/pkg/trading"
)

type emptyBody struct{}

type stubCommand struct {
	callName string
	body     interface{}
	response trading.Response
	parseErr error
}

func (c stubCommand) CallName() string                                 { return c.callName }
func (c stubCommand) Body() interface{}                                { return c.body }
func (c stubCommand) ParseResponse(r []byte) (trading.Response, error) { return c.response, c.parseErr }

func TestRunCommandSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "text/xml" {
			t.Errorf("Content-Type = %q", ct)
		}
		if cn := r.Header.Get("X-EBAY-API-CALL-NAME"); cn != "TestCall" {
			t.Errorf("X-EBAY-API-CALL-NAME = %q", cn)
		}
		if dev := r.Header.Get("X-EBAY-API-DEV-NAME"); dev != "dev123" {
			t.Errorf("X-EBAY-API-DEV-NAME = %q", dev)
		}
		if app := r.Header.Get("X-EBAY-API-APP-NAME"); app != "app123" {
			t.Errorf("X-EBAY-API-APP-NAME = %q", app)
		}
		if cert := r.Header.Get("X-EBAY-API-CERT-NAME"); cert != "cert123" {
			t.Errorf("X-EBAY-API-CERT-NAME = %q", cert)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	conf := trading.Conf{
		BaseURL:   server.URL,
		DevId:     "dev123",
		AppId:     "app123",
		CertId:    "cert123",
		AuthToken: "token123",
		SiteId:    0,
	}

	cmd := stubCommand{
		callName: "TestCall",
		body:     emptyBody{},
		response: trading.GenericResponse{Ack: "Success"},
	}

	resp, err := conf.RunCommand(cmd)
	if err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if resp.Failure() {
		t.Error("expected success")
	}
}

func TestRunCommandHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "internal server error")
	}))
	defer server.Close()

	conf := trading.Conf{BaseURL: server.URL, DevId: "d", AppId: "a", CertId: "c", AuthToken: "t"}
	cmd := stubCommand{callName: "TestCall", body: emptyBody{}, response: trading.GenericResponse{}}
	if _, err := conf.RunCommand(cmd); err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestRunCommandFailureAck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Failure</Ack><Errors><ShortMessage>Bad</ShortMessage><LongMessage>Bad request</LongMessage><ErrorCode>100</ErrorCode></Errors></TestCallResponse>`)
	}))
	defer server.Close()

	conf := trading.Conf{BaseURL: server.URL, DevId: "d", AppId: "a", CertId: "c", AuthToken: "t"}
	cmd := stubCommand{callName: "TestCall", body: emptyBody{}, response: trading.GenericResponse{Ack: "Failure"}}

	resp, err := conf.RunCommand(cmd)
	if err == nil {
		t.Fatal("expected error for Failure ack")
	}
	if !resp.Failure() {
		t.Error("expected Failure() == true")
	}
}

func TestSandboxAndProduction(t *testing.T) {
	conf := trading.Conf{}
	if got := conf.Sandbox().BaseURL; got != "https://api.sandbox.ebay.com" {
		t.Errorf("Sandbox BaseURL = %q", got)
	}
	if got := conf.Production().BaseURL; got != "https://api.ebay.com" {
		t.Errorf("Production BaseURL = %q", got)
	}
}

func TestRunCommandSendsAuthToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), "mySecretToken") {
			t.Error("body missing auth token")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	conf := trading.Conf{BaseURL: server.URL, DevId: "d", AppId: "a", CertId: "c", AuthToken: "mySecretToken"}
	cmd := stubCommand{callName: "TestCall", body: emptyBody{}, response: trading.GenericResponse{Ack: "Success"}}
	if _, err := conf.RunCommand(cmd); err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
}

func TestRunCommandLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<TestCallResponse><Ack>Success</Ack></TestCallResponse>`)
	}))
	defer server.Close()

	logCalls := 0
	conf := trading.Conf{
		BaseURL: server.URL, DevId: "d", AppId: "a", CertId: "c", AuthToken: "t",
		Logger: func(args ...interface{}) { logCalls++ },
	}
	cmd := stubCommand{callName: "TestCall", body: emptyBody{}, response: trading.GenericResponse{Ack: "Success"}}
	if _, err := conf.RunCommand(cmd); err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	if logCalls != 2 {
		t.Errorf("Logger called %d times, want 2", logCalls)
	}
}

func TestTimestampUnmarshal(t *testing.T) {
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
				T trading.Timestamp `xml:"T"`
			}
			if err := xml.Unmarshal([]byte(`<R>`+tt.xml+`</R>`), &v); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			if got := v.T.Format("2006-01-02"); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("invalid format", func(t *testing.T) {
		var v struct {
			T trading.Timestamp `xml:"T"`
		}
		if err := xml.Unmarshal([]byte(`<R><T>not-a-date</T></R>`), &v); err == nil {
			t.Error("expected error for invalid date format")
		}
	})
}
