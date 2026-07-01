package postmark

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/comms"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, APIKey: "server-token", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

// serverToken is a client whose base points at a server that must never be hit;
// used to prove request-level validation happens before any HTTP call.
func mustNotCall(t *testing.T) *Client {
	t.Helper()
	return newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("HTTP call made despite invalid input: %s %s", r.Method, r.URL.Path)
	})
}

func TestSend_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/email" {
			t.Errorf("path = %q, want /email", r.URL.Path)
		}
		if got := r.Header.Get(authHeader); got != "server-token" {
			t.Errorf("%s = %q, want server-token", authHeader, got)
		}
		var req sendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.From != "from@x.com" || req.To != "to@x.com" || req.Subject != "Hi" {
			t.Errorf("unexpected payload: %+v", req)
		}
		if req.HtmlBody != "<b>hi</b>" {
			t.Errorf("HtmlBody = %q", req.HtmlBody)
		}
		if req.MessageStream != "outbound" {
			t.Errorf("MessageStream = %q, want outbound", req.MessageStream)
		}
		_, _ = w.Write([]byte(`{"ErrorCode":0,"MessageID":"abc","Message":"OK"}`))
	})

	err := c.Send(context.Background(), comms.Email{
		From: "from@x.com", To: "to@x.com", Subject: "Hi", HTML: "<b>hi</b>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_MissingFrom(t *testing.T) {
	c := mustNotCall(t)
	err := c.Send(context.Background(), comms.Email{To: "to@x.com", HTML: "hi"})
	if err == nil || !strings.Contains(err.Error(), "From is required") {
		t.Fatalf("err = %v, want From-required error", err)
	}
}

func TestSend_MissingBody(t *testing.T) {
	c := mustNotCall(t)
	err := c.Send(context.Background(), comms.Email{From: "from@x.com", To: "to@x.com"})
	if err == nil || !strings.Contains(err.Error(), "HTML or Text") {
		t.Fatalf("err = %v, want body-required error", err)
	}
}

func TestSend_RejectedByErrorCode(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// HTTP 200 but a non-zero ErrorCode means Postmark rejected the message.
		_, _ = w.Write([]byte(`{"ErrorCode":300,"Message":"Invalid email request"}`))
	})
	err := c.Send(context.Background(), comms.Email{From: "f@x.com", To: "t@x.com", Text: "hi"})
	if err == nil || !strings.Contains(err.Error(), "300") {
		t.Fatalf("err = %v, want rejection with code 300", err)
	}
}
