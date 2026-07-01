package transgov

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"car-bridge/internal/integrations"
)

// vPIC is a public API, so the client is built with an empty APIKey and must not
// send an Authorization header.
func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func TestDecode_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/DecodeVinValues/1HGCM82633A004352" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("format"); got != "json" {
			t.Errorf("format = %q, want json", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("Authorization = %q, want empty (public API)", got)
		}
		_, _ = w.Write([]byte(`{
			"Count": 1,
			"Results": [{
				"VIN": "1HGCM82633A004352",
				"Make": "HONDA",
				"Model": "Accord",
				"ModelYear": "2003",
				"BodyClass": "Sedan",
				"FuelTypePrimary": "Gasoline"
			}]
		}`))
	})

	spec, err := c.Decode(context.Background(), "1HGCM82633A004352")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if spec.Make != "HONDA" || spec.Model != "Accord" || spec.ModelYear != 2003 {
		t.Errorf("unexpected spec: %+v", spec)
	}
	if spec.BodyClass != "Sedan" || spec.FuelType != "Gasoline" {
		t.Errorf("unexpected body/fuel: %+v", spec)
	}
}

func TestDecode_EmptyResultsIsNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Count":0,"Results":[]}`))
	})
	_, err := c.Decode(context.Background(), "BADVIN")
	if err != integrations.ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDecode_VINFallsBackToInput(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// vPIC sometimes returns a result with an empty VIN field.
		_, _ = w.Write([]byte(`{"Count":1,"Results":[{"Make":"HONDA","ModelYear":"2003"}]}`))
	})
	spec, err := c.Decode(context.Background(), "MYVIN123")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if spec.VIN != "MYVIN123" {
		t.Errorf("VIN = %q, want the input VIN MYVIN123", spec.VIN)
	}
}

func TestDecode_UnparseableYearIsZero(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Count":1,"Results":[{"VIN":"X","ModelYear":"n/a"}]}`))
	})
	spec, err := c.Decode(context.Background(), "X")
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if spec.ModelYear != 0 {
		t.Errorf("ModelYear = %d, want 0 for unparseable year", spec.ModelYear)
	}
}
