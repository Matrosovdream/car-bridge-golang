// Package carsxe implements vehicle.PlateDecoder using the CarsXE API
// (license-plate-to-vehicle lookup).
package carsxe

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/vehicle"
)

// Client decodes license plates via the CarsXE v2 plate decoder.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a carsxe client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ vehicle.PlateDecoder = (*Client)(nil)

// flexInt accepts either a JSON number or a numeric string — CarsXE returns the
// year as one or the other depending on the source country.
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	if n, err := strconv.Atoi(s); err == nil {
		*f = flexInt(n)
	}
	return nil
}

// plateResponse is the CarsXE v2 /platedecoder reply (a flat object).
type plateResponse struct {
	Success          bool    `json:"success"`
	VIN              string  `json:"vin"`
	Make             string  `json:"make"`
	Model            string  `json:"model"`
	Year             flexInt `json:"year"`
	RegistrationYear flexInt `json:"registration_year"`
	BodyStyle        string  `json:"body_style"`
	FuelType         string  `json:"fuel_type"`
}

// DecodePlate resolves a license plate (+ state) to vehicle specs. CarsXE
// authenticates with a `key` query param.
func (c *Client) DecodePlate(ctx context.Context, plate, state string) (*vehicle.VehicleSpec, error) {

	q := url.Values{}
	q.Set("key", c.cfg.APIKey)
	q.Set("plate", plate)
	q.Set("state", state)
	q.Set("country", "US")

	var raw plateResponse
	if err := c.base.DoJSON(
		ctx, http.MethodGet,
		c.base.URL("v2/platedecoder?"+q.Encode()),
		nil, &raw, nil,
	); err != nil {
		return nil, err
	}
	if !raw.Success || (raw.Make == "" && raw.VIN == "") {
		return nil, integrations.ErrNotFound
	}

	year := int(raw.Year)
	if year == 0 {
		year = int(raw.RegistrationYear)
	}
	return &vehicle.VehicleSpec{
		VIN:       raw.VIN,
		Make:      raw.Make,
		Model:     raw.Model,
		ModelYear: year,
		BodyClass: raw.BodyStyle,
		FuelType:  raw.FuelType,
	}, nil

}
