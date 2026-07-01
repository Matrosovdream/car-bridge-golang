package carsxe

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
	cfg := integrations.HTTPConfig{BaseURL: srv.URL, APIKey: "test-key", TimeoutSeconds: 5}
	return New(cfg, integrations.NewClient(cfg, nil))
}

func TestDecodePlate_Success(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/platedecoder" {
			t.Errorf("path = %q, want /v2/platedecoder", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("key") != "test-key" {
			t.Errorf("key = %q, want test-key", q.Get("key"))
		}
		if q.Get("plate") != "ABC123" {
			t.Errorf("plate = %q, want ABC123", q.Get("plate"))
		}
		if q.Get("state") != "CA" {
			t.Errorf("state = %q, want CA", q.Get("state"))
		}
		if q.Get("country") != "US" {
			t.Errorf("country = %q, want US", q.Get("country"))
		}
		_, _ = w.Write([]byte(`{
			"success": true,
			"vin": "1HGCM82633A004352",
			"make": "Honda",
			"model": "Accord",
			"year": 2003,
			"body_style": "Sedan",
			"fuel_type": "Gasoline"
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
	if spec.BodyClass != "Sedan" || spec.FuelType != "Gasoline" {
		t.Errorf("unexpected body/fuel: %+v", spec)
	}
}

func TestDecodePlate_YearAsString(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"make":"Ford","year":"2019"}`))
	})
	spec, err := c.DecodePlate(context.Background(), "P", "TX")
	if err != nil {
		t.Fatalf("DecodePlate: %v", err)
	}
	if spec.ModelYear != 2019 {
		t.Errorf("ModelYear = %d, want 2019 (string year)", spec.ModelYear)
	}
}

func TestDecodePlate_YearFallsBackToRegistrationYear(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"make":"Ford","year":0,"registration_year":2018}`))
	})
	spec, err := c.DecodePlate(context.Background(), "P", "TX")
	if err != nil {
		t.Fatalf("DecodePlate: %v", err)
	}
	if spec.ModelYear != 2018 {
		t.Errorf("ModelYear = %d, want 2018 (registration_year fallback)", spec.ModelYear)
	}
}

func TestDecodePlate_NotFound(t *testing.T) {
	cases := map[string]string{
		"success false":     `{"success":false,"make":"Honda"}`,
		"empty make and vin": `{"success":true}`,
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

func TestFlexInt_Unmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want flexInt
	}{
		{`123`, 123},
		{`"123"`, 123},
		{`""`, 0},
		{`"null"`, 0},
		{`null`, 0},
		{`"abc"`, 0}, // unparseable strings are ignored, not errors
	}
	for _, tc := range cases {
		var f flexInt
		if err := f.UnmarshalJSON([]byte(tc.in)); err != nil {
			t.Errorf("UnmarshalJSON(%s) error: %v", tc.in, err)
		}
		if f != tc.want {
			t.Errorf("UnmarshalJSON(%s) = %d, want %d", tc.in, f, tc.want)
		}
	}
}
