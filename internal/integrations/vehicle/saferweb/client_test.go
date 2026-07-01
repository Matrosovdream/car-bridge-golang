package saferweb

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
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, APIKey: "web-key", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func TestLookupByDOT_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/carriers/123456" {
			t.Errorf("path = %q, want /carriers/123456", r.URL.Path)
		}
		if got := r.URL.Query().Get("webKey"); got != "web-key" {
			t.Errorf("webKey = %q, want web-key", got)
		}
		_, _ = w.Write([]byte(`{
			"content": {
				"carrier": {
					"dotNumber": 123456,
					"legalName": "ACME TRUCKING LLC",
					"dbaName": "ACME",
					"phyState": "TX",
					"statusCode": "A",
					"totalPowerUnits": 12,
					"totalDrivers": 20
				}
			}
		}`))
	})

	car, err := c.LookupByDOT(context.Background(), "123456")
	if err != nil {
		t.Fatalf("LookupByDOT: %v", err)
	}
	if car.DOTNumber != "123456" {
		t.Errorf("DOTNumber = %q, want 123456", car.DOTNumber)
	}
	if car.LegalName != "ACME TRUCKING LLC" || car.DBAName != "ACME" {
		t.Errorf("unexpected names: %+v", car)
	}
	if car.PhyState != "TX" || car.Status != "A" {
		t.Errorf("unexpected state/status: %+v", car)
	}
	if car.TotalTrucks != 12 || car.TotalDrivers != 20 {
		t.Errorf("unexpected counts: %+v", car)
	}
	// EntityType is intentionally unset: QCMobile has no direct field.
	if car.EntityType != "" {
		t.Errorf("EntityType = %q, want empty", car.EntityType)
	}
}

func TestLookupByDOT_NotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Empty carrier envelope: zero dotNumber and no legal name.
		_, _ = w.Write([]byte(`{"content":{"carrier":{}}}`))
	})
	_, err := c.LookupByDOT(context.Background(), "999")
	if err != integrations.ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestLookupByDOT_Upstream404(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err := c.LookupByDOT(context.Background(), "999")
	if err != integrations.ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
