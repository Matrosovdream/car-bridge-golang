package transgov

import (
	"context"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/vehicle"
)

type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

func New(
	cfg integrations.HTTPConfig,
	base *integrations.Client,
) *Client {

	return &Client{
		base: base,
		cfg:  cfg,
	}

}

var _ vehicle.VINDecoder = (*Client)(nil)

func (c *Client) Decode(
	ctx context.Context, vin string,
) (*vehicle.VehicleSpec, error) {

	// Implement later

	return nil, integrations.ErrNotImplemented

}
