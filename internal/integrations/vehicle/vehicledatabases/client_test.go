package vehicledatabases

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"car-bridge/internal/integrations"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, APIKey: "auth-key", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func TestDecodePlate_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/license-decode/ABC123/CA" {
			t.Errorf("path = %q, want /license-decode/ABC123/CA", r.URL.Path)
		}
		if got := r.Header.Get(authHeader); got != "auth-key" {
			t.Errorf("%s = %q, want auth-key", authHeader, got)
		}
		_, _ = w.Write([]byte(`{
			"status": "success",
			"data": {
				"intro": {"vin": "1HGCM82633A004352"},
				"basic": {"make": "Honda", "model": "Accord", "year": 2003}
			}
		}`))
	})

	spec, err := c.DecodePlate(context.Background(), "ABC123", "CA")
	if err != nil {
		t.Fatalf("DecodePlate: %v", err)
	}
	if spec.VIN != "1HGCM82633A004352" || spec.Make != "Honda" || spec.Model != "Accord" {
		t.Errorf("unexpected spec: %+v", spec)
	}
	if spec.ModelYear != 2003 {
		t.Errorf("ModelYear = %d, want 2003", spec.ModelYear)
	}
}

func TestDecodePlate_NotFound(t *testing.T) {
	cases := map[string]string{
		"status not success": `{"status":"error"}`,
		"empty make and vin": `{"status":"success","data":{}}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(body))
			})
			_, err := c.DecodePlate(context.Background(), "P", "TX")
			if err != integrations.ErrNotFound {
				t.Fatalf("err = %v, want ErrNotFound", err)
			}
		})
	}
}
