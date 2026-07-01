package twilio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/comms"
)

// newTestClient builds a client whose base URL is the test server plus the given
// suffix (Twilio's account SID lives in the base URL path).
func newTestClient(t *testing.T, suffix string, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := integrations.HTTPConfig{BaseURL: srv.URL + suffix, APIKey: "auth-token", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func TestSend_Success(t *testing.T) {
	c := newTestClient(t, "/Accounts/AC123", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/Messages.json") {
			t.Errorf("path = %q, want suffix /Messages.json", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "AC123" || pass != "auth-token" {
			t.Errorf("basic auth = (%q,%q,%v), want (AC123, auth-token, true)", user, pass, ok)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("To") != "+15551112222" || r.FormValue("From") != "+15553334444" {
			t.Errorf("unexpected to/from: To=%q From=%q", r.FormValue("To"), r.FormValue("From"))
		}
		if r.FormValue("Body") != "hello" {
			t.Errorf("Body = %q, want hello", r.FormValue("Body"))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"sid":"SM1","status":"queued"}`))
	})

	err := c.Send(context.Background(), comms.Message{
		From: "+15553334444", To: "+15551112222", Body: "hello",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestSend_MissingAccountSID(t *testing.T) {
	c := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP call made despite missing account SID")
	})
	err := c.Send(context.Background(), comms.Message{From: "+1", To: "+2", Body: "x"})
	if err == nil || !strings.Contains(err.Error(), "account SID") {
		t.Fatalf("err = %v, want account-SID error", err)
	}
}

func TestSend_MissingFrom(t *testing.T) {
	c := newTestClient(t, "/Accounts/AC123", func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP call made despite missing From")
	})
	err := c.Send(context.Background(), comms.Message{To: "+2", Body: "x"})
	if err == nil || !strings.Contains(err.Error(), "From is required") {
		t.Fatalf("err = %v, want From-required error", err)
	}
}

func TestSend_ErrorMapping(t *testing.T) {
	t.Run("rate limited", func(t *testing.T) {
		c := newTestClient(t, "/Accounts/AC123", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})
		err := c.Send(context.Background(), comms.Message{From: "+1", To: "+2", Body: "x"})
		if err != integrations.ErrRateLimited {
			t.Fatalf("err = %v, want ErrRateLimited", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		c := newTestClient(t, "/Accounts/AC123", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		err := c.Send(context.Background(), comms.Message{From: "+1", To: "+2", Body: "x"})
		if err == nil || !strings.Contains(err.Error(), "upstream unavailable") {
			t.Fatalf("err = %v, want upstream-unavailable error", err)
		}
	})

	t.Run("client error with json detail", func(t *testing.T) {
		c := newTestClient(t, "/Accounts/AC123", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":21211,"message":"Invalid 'To' Phone Number"}`))
		})
		err := c.Send(context.Background(), comms.Message{From: "+1", To: "bad", Body: "x"})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "21211") || !strings.Contains(err.Error(), "Invalid 'To'") {
			t.Fatalf("err = %v, want code+message detail", err)
		}
	})
}

func TestAccountSIDFromURL(t *testing.T) {
	cases := map[string]string{
		"https://api.twilio.com/2010-04-01/Accounts/AC123":  "AC123",
		"https://api.twilio.com/2010-04-01/Accounts/AC123/": "AC123",
		"https://api.twilio.com/2010-04-01":                 "",
		"":                                                  "",
	}
	for in, want := range cases {
		if got := accountSIDFromURL(in); got != want {
			t.Errorf("accountSIDFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}
