// Package saferweb integrates FMCSA SAFER / QCMobile for motor-carrier safety data.
// QCMobile has a history of downtime, so the shared client retries with backoff.
package saferweb

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/vehicle"
)

// Client looks up carriers by DOT number via FMCSA QCMobile.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a saferweb client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ vehicle.CarrierLookup = (*Client)(nil)

// carrierResponse is the QCMobile /carriers/{dot} envelope.
type carrierResponse struct {
	Content struct {
		Carrier struct {
			DotNumber       int    `json:"dotNumber"`
			LegalName       string `json:"legalName"`
			DbaName         string `json:"dbaName"`
			PhyState        string `json:"phyState"`
			StatusCode      string `json:"statusCode"`
			TotalPowerUnits int    `json:"totalPowerUnits"`
			TotalDrivers    int    `json:"totalDrivers"`
		} `json:"carrier"`
	} `json:"content"`
}

// LookupByDOT resolves carrier data for a USDOT number. QCMobile authenticates
// with a webKey query param (not a header), so the key is added to the URL.
func (c *Client) LookupByDOT(ctx context.Context, dotNumber string) (*vehicle.Carrier, error) {

	path := "carriers/" + url.PathEscape(dotNumber) + "?webKey=" + url.QueryEscape(c.cfg.APIKey)

	var raw carrierResponse
	if err := c.base.DoJSON(ctx, http.MethodGet, c.base.URL(path), nil, &raw, nil); err != nil {
		return nil, err
	}

	cr := raw.Content.Carrier
	if cr.DotNumber == 0 && cr.LegalName == "" {
		return nil, integrations.ErrNotFound
	}

	// EntityType is intentionally left empty: QCMobile does not expose a direct
	// SAFER "entity type" field.
	return &vehicle.Carrier{
		DOTNumber:    strconv.Itoa(cr.DotNumber),
		LegalName:    cr.LegalName,
		DBAName:      cr.DbaName,
		PhyState:     cr.PhyState,
		Status:       cr.StatusCode,
		TotalTrucks:  cr.TotalPowerUnits,
		TotalDrivers: cr.TotalDrivers,
	}, nil

}
