package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newServer starts an httptest server and returns a base Client pointed at it.
// APIKey defaults to "secret" and Retries to 0 so tests don't pay backoff unless
// they opt in.
func newServer(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := NewClient(HTTPConfig{BaseURL: srv.URL, APIKey: "secret", TimeoutSeconds: 5}, nil)
	return c, srv
}

func TestURL(t *testing.T) {
	cases := []struct {
		base, path, want string
	}{
		{"http://x", "y", "http://x/y"},
		{"http://x/", "/y", "http://x/y"},
		{"http://x/", "y", "http://x/y"},
		{"http://x", "/y", "http://x/y"},
		{"http://x/api/", "/v2/z", "http://x/api/v2/z"},
	}
	for _, tc := range cases {
		c := NewClient(HTTPConfig{BaseURL: tc.base}, nil)
		if got := c.URL(tc.path); got != tc.want {
			t.Errorf("URL(%q,%q) = %q, want %q", tc.base, tc.path, got, tc.want)
		}
	}
}

func TestNewClient_Timeout(t *testing.T) {
	cases := []struct {
		seconds int
		want    time.Duration
	}{
		{0, 10 * time.Second},
		{-3, 10 * time.Second},
		{7, 7 * time.Second},
	}
	for _, tc := range cases {
		c := NewClient(HTTPConfig{TimeoutSeconds: tc.seconds}, nil)
		if c.http.Timeout != tc.want {
			t.Errorf("TimeoutSeconds=%d -> %v, want %v", tc.seconds, c.http.Timeout, tc.want)
		}
	}
}

func TestDoJSON_Success_DecodesAndSetsHeaders(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Errorf("Authorization = %q, want Bearer secret", got)
		}
		// GET with no body must not advertise a content type.
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("Content-Type = %q, want empty for GET", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.DoJSON(context.Background(), http.MethodGet, c.URL("/x"), nil, &out, nil); err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true, got %+v", out)
	}
}

func TestDoJSON_PostMarshalsBodyAndSetsContentType(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var in struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if in.Name != "hello" {
			t.Errorf("body name = %q, want hello", in.Name)
		}
		w.WriteHeader(http.StatusOK)
	})

	body := map[string]string{"name": "hello"}
	if err := c.DoJSON(context.Background(), http.MethodPost, c.URL("/x"), body, nil, nil); err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
}

func TestDoJSON_CustomHeadersOverrideDefaults(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Custom"); got != "v" {
			t.Errorf("X-Custom = %q, want v", got)
		}
		// A caller-supplied Authorization must win over the Bearer default.
		if got := r.Header.Get("Authorization"); got != "Token xyz" {
			t.Errorf("Authorization = %q, want Token xyz", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	headers := map[string]string{"X-Custom": "v", "Authorization": "Token xyz"}
	if err := c.DoJSON(context.Background(), http.MethodGet, c.URL("/x"), nil, nil, headers); err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
}

func TestDoJSON_EmptyBodyIsNotAnError(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // 200, no body
	})
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.DoJSON(context.Background(), http.MethodGet, c.URL("/x"), nil, &out, nil); err != nil {
		t.Fatalf("DoJSON: %v", err)
	}
	if out.OK {
		t.Fatalf("out should be untouched on empty body, got %+v", out)
	}
}

func TestDoJSON_StatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		wantErr    error  // sentinel expected via errors.Is (nil = match by substring)
		wantSubstr string // substring expected when wantErr is nil
	}{
		{"not found", http.StatusNotFound, ErrNotFound, ""},
		{"rate limited", http.StatusTooManyRequests, ErrRateLimited, ""},
		{"server error", http.StatusInternalServerError, ErrUpstreamUnavailable, ""},
		{"bad gateway", http.StatusBadGateway, ErrUpstreamUnavailable, ""},
		{"unexpected 400", http.StatusBadRequest, nil, "unexpected status 400"},
		{"unexpected 403", http.StatusForbidden, nil, "unexpected status 403"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			})
			err := c.DoJSON(context.Background(), http.MethodGet, c.URL("/x"), nil, nil, nil)
			if err == nil {
				t.Fatalf("expected error for status %d", tc.status)
			}
			if tc.wantErr != nil && !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want errors.Is %v", err, tc.wantErr)
			}
			if tc.wantSubstr != "" && !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Fatalf("err = %v, want substring %q", err, tc.wantSubstr)
			}
		})
	}
}

func TestDoJSON_BadResponseBody(t *testing.T) {
	c, _ := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{not json`))
	})
	var out struct {
		OK bool `json:"ok"`
	}
	err := c.DoJSON(context.Background(), http.MethodGet, c.URL("/x"), nil, &out, nil)
	if !errors.Is(err, ErrBadResponse) {
		t.Fatalf("err = %v, want errors.Is ErrBadResponse", err)
	}
}

func TestDoJSON_RetriesThenSucceeds(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient(HTTPConfig{BaseURL: srv.URL, TimeoutSeconds: 5, Retries: 2}, nil)

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.DoJSON(context.Background(), http.MethodGet, srv.URL, nil, &out, nil); err != nil {
		t.Fatalf("DoJSON returned error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true, got %+v", out)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls (1 retry), got %d", calls)
	}
}

func TestDoJSON_RetriesExhaustedReturnsLastErr(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient(HTTPConfig{BaseURL: srv.URL, TimeoutSeconds: 5, Retries: 1}, nil)
	err := c.DoJSON(context.Background(), http.MethodGet, srv.URL, nil, nil, nil)
	if !errors.Is(err, ErrUpstreamUnavailable) {
		t.Fatalf("err = %v, want ErrUpstreamUnavailable", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls (1 retry), got %d", calls)
	}
}

func TestDoJSON_NonRetryableDoesNotRetry(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := NewClient(HTTPConfig{BaseURL: srv.URL, TimeoutSeconds: 5, Retries: 3}, nil)
	if err := c.DoJSON(context.Background(), http.MethodGet, srv.URL, nil, nil, nil); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("400 must not retry: expected 1 call, got %d", calls)
	}
}

func TestDoJSON_ContextCanceledDuringBackoff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // always retryable
	}))
	defer srv.Close()

	c := NewClient(HTTPConfig{BaseURL: srv.URL, TimeoutSeconds: 5, Retries: 5}, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := c.DoJSON(ctx, http.MethodGet, srv.URL, nil, nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestDoJSON_NetworkErrorIsUpstreamUnavailable(t *testing.T) {
	// Point at a closed server so Do() fails at the transport layer.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := NewClient(HTTPConfig{BaseURL: url, TimeoutSeconds: 1}, nil)
	err := c.DoJSON(context.Background(), http.MethodGet, url, nil, nil, nil)
	if !errors.Is(err, ErrUpstreamUnavailable) {
		t.Fatalf("err = %v, want ErrUpstreamUnavailable", err)
	}
}

func TestIsRetryable(t *testing.T) {
	if !isRetryable(ErrUpstreamUnavailable) {
		t.Error("ErrUpstreamUnavailable should be retryable")
	}
	if !isRetryable(ErrRateLimited) {
		t.Error("ErrRateLimited should be retryable")
	}
	if isRetryable(ErrNotFound) {
		t.Error("ErrNotFound should not be retryable")
	}
	if isRetryable(errors.New("plain")) {
		t.Error("plain error should not be retryable")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate([]byte("short")); got != "short" {
		t.Errorf("truncate(short) = %q", got)
	}
	long := strings.Repeat("a", 300)
	got := truncate([]byte(long))
	if len(got) == len(long) {
		t.Fatal("long input was not truncated")
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("truncated output should end with ellipsis, got %q", got)
	}
	if !strings.HasPrefix(got, strings.Repeat("a", 256)) {
		t.Error("truncated output should keep the first 256 bytes")
	}
}
