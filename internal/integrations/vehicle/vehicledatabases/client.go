// Package vehicledatabases implements vehicle.PlateDecoder using the
// Vehicle Databases license-plate decode API.
package vehicledatabases

import (
	"context"
	"net/http"
	"net/url"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/vehicle"
)

// authHeader is the Vehicle Databases API-key header.
const authHeader = "x-authkey"

// Client decodes license plates via the Vehicle Databases API.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a vehicledatabases client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ vehicle.PlateDecoder = (*Client)(nil)

// plateResponse is the Vehicle Databases /license-decode reply.
type plateResponse struct {
	Status string `json:"status"`
	Data   struct {
		Intro struct {
			VIN string `json:"vin"`
		} `json:"intro"`
		Basic struct {
			Make  string `json:"make"`
			Model string `json:"model"`
			Year  int    `json:"year"`
		} `json:"basic"`
	} `json:"data"`
}

// DecodePlate resolves a license plate (+ state) to vehicle specs. The plate and
// state are path segments; auth is the x-authkey header.
func (c *Client) DecodePlate(ctx context.Context, plate, state string) (*vehicle.VehicleSpec, error) {

	path := "license-decode/" + url.PathEscape(plate) + "/" + url.PathEscape(state)
	headers := map[string]string{authHeader: c.cfg.APIKey}

	var raw plateResponse
	if err := c.base.DoJSON(
		ctx, http.MethodGet,
		c.base.URL(path), nil, &raw, headers,
	); err != nil {
		return nil, err
	}

	b := raw.Data.Basic
	if raw.Status != "success" || (b.Make == "" && raw.Data.Intro.VIN == "") {
		return nil, integrations.ErrNotFound
	}

	return &vehicle.VehicleSpec{
		VIN:       raw.Data.Intro.VIN,
		Make:      b.Make,
		Model:     b.Model,
		ModelYear: b.Year,
	}, nil

}
