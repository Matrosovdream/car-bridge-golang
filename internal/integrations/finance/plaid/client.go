// Package plaid integrates Plaid for bank/income verification.
package plaid

import (
	"context"

	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/finance"
)

// Client reads bank accounts via the Plaid API.
type Client struct {
	base *integrations.Client
	cfg  integrations.HTTPConfig
}

// New builds a plaid client from config and a shared base HTTP client.
func New(cfg integrations.HTTPConfig, base *integrations.Client) *Client {
	return &Client{base: base, cfg: cfg}
}

// compile-time check that Client satisfies the port.
var _ finance.BankVerifier = (*Client)(nil)

// Accounts returns the bank accounts behind an access token.
func (c *Client) Accounts(ctx context.Context, accessToken string) ([]finance.BankAccount, error) {
	// TODO: POST {base_url}/accounts/get with {client_id, secret, access_token}
	return nil, integrations.ErrNotImplemented
}
