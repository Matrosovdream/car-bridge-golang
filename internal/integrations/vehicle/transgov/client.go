// Package transgov integrates DOT/NHTSA vPIC (+ FMCSA datasets) for vehicle specs.
package transgov

import (
	"context"
	"net/http"
	"strconv"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/vehicle"
)

// Client decodes VINs via the NHTSA vPIC API.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a transgov client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ vehicle.VINDecoder = (*Client)(nil)

// decodeResponse is the vPIC DecodeVinValues envelope. DecodeVinValues returns a
// single flat result object with every decoded variable as a field.
type decodeResponse struct {
	Count   int            `json:"Count"`
	Message string         `json:"Message"`
	Results []decodeResult `json:"Results"`
}

type decodeResult struct {
	VIN             string `json:"VIN"`
	Make            string `json:"Make"`
	Model           string `json:"Model"`
	ModelYear       string `json:"ModelYear"`
	BodyClass       string `json:"BodyClass"`
	FuelTypePrimary string `json:"FuelTypePrimary"`
}

// Decode resolves a VIN to vehicle specs. vPIC is public, so no auth is sent.
func (c *Client) Decode(ctx context.Context, vin string) (*vehicle.VehicleSpec, error) {

	url := c.base.URL("DecodeVinValues/" + vin + "?format=json")

	var raw decodeResponse
	if err := c.base.DoJSON(ctx, http.MethodGet, url, nil, &raw, nil); err != nil {
		return nil, err
	}
	if len(raw.Results) == 0 {
		return nil, integrations.ErrNotFound
	}

	r := raw.Results[0]
	// vPIC reports ModelYear as a string; ignore unparseable values.
	year, _ := strconv.Atoi(r.ModelYear)

	out := &vehicle.VehicleSpec{
		VIN:       r.VIN,
		Make:      r.Make,
		Model:     r.Model,
		ModelYear: year,
		BodyClass: r.BodyClass,
		FuelType:  r.FuelTypePrimary,
	}
	if out.VIN == "" {
		out.VIN = vin
	}
	return out, nil

}
