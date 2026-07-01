package sendgrid

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
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, APIKey: "sg-key", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func mustNotCall(t *testing.T) *Client {
	t.Helper()
	return newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("HTTP call made despite invalid input: %s %s", r.Method, r.URL.Path)
	})
}

func TestSend_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/mail/send" {
			t.Errorf("got %s %s, want POST /v3/mail/send", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sg-key" {
			t.Errorf("Authorization = %q, want Bearer sg-key", got)
		}
		var req sendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.From.Email != "from@x.com" || req.Subject != "Hi" {
			t.Errorf("unexpected payload: %+v", req)
		}
		if len(req.Personalizations) != 1 || len(req.Personalizations[0].To) != 1 ||
			req.Personalizations[0].To[0].Email != "to@x.com" {
			t.Errorf("unexpected personalizations: %+v", req.Personalizations)
		}
		// SendGrid accepts with 202 and an empty body.
		w.WriteHeader(http.StatusAccepted)
	})

	err := c.Send(context.Background(), comms.Email{
		From: "from@x.com", To: "to@x.com", Subject: "Hi", HTML: "<b>hi</b>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_ContentOrderTextBeforeHTML(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req sendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Content) != 2 {
			t.Fatalf("content len = %d, want 2", len(req.Content))
		}
		// SendGrid requires text/plain to precede text/html.
		if req.Content[0].Type != "text/plain" || req.Content[1].Type != "text/html" {
			t.Errorf("content order = [%s,%s], want [text/plain,text/html]",
				req.Content[0].Type, req.Content[1].Type)
		}
		w.WriteHeader(http.StatusAccepted)
	})

	err := c.Send(context.Background(), comms.Email{
		From: "f@x.com", To: "t@x.com", Subject: "S", Text: "plain", HTML: "<b>rich</b>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_MissingFrom(t *testing.T) {
	c := mustNotCall(t)
	err := c.Send(context.Background(), comms.Email{To: "t@x.com", HTML: "hi"})
	if err == nil || !strings.Contains(err.Error(), "From is required") {
		t.Fatalf("err = %v, want From-required error", err)
	}
}

func TestSend_MissingBody(t *testing.T) {
	c := mustNotCall(t)
	err := c.Send(context.Background(), comms.Email{From: "f@x.com", To: "t@x.com"})
	if err == nil || !strings.Contains(err.Error(), "HTML or Text") {
		t.Fatalf("err = %v, want body-required error", err)
	}
}
