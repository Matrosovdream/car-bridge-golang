package integrations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSON_RetriesThenSuccedes(t *testing.T) {

	var calls int

	srv := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter, r *http.Request,
			) {

				calls++
				if calls == 1 {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"ok":true}`),
				)

			},
		))
	defer srv.Close()

	c := NewClient(
		HTTPConfig{
			BaseURL:        srv.URL,
			TimeoutSeconds: 5,
			Retries:        2,
		}, nil,
	)

	var out struct {
		OK bool `json:"ok"`
	}

	if err := c.DoJSON(
		context.Background(),
		http.MethodGet,
		srv.URL, nil, &out, nil,
	); err != nil {
		t.Fatalf("DoJSON returned error: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true, got %+v", out)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls (1 retry), got %d", calls)
	}

}
