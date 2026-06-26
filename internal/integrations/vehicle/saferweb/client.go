package saferweb

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
	cfg integrations.HTTPConfig, base *integrations.Client,
) *Client {

	return &Client{
		base: base,
		cfg:  cfg,
	}

}

var _ vehicle.CarrierLookup = (*Client)(nil)

func (c *Client) LookupByDOT(
	ctx context.Context, dotnumber string,
) (*vehicle.Carrier, error) {

	// Implement later

	return nil, integrations.ErrNotImplemented

}
