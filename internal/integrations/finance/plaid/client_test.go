package plaid

import (
	"context"
	"errors"
	"testing"

	"car-bridge/internal/integrations"
)

// Accounts is a documented stub for now; lock in the ErrNotImplemented contract
// so wiring it up later is a deliberate change.
func TestAccounts_NotImplemented(t *testing.T) {
	c := New(integrations.HTTPConfig{}, integrations.NewClient(integrations.HTTPConfig{}, nil))
	accounts, err := c.Accounts(context.Background(), "access-token")
	if !errors.Is(err, integrations.ErrNotImplemented) {
		t.Fatalf("err = %v, want ErrNotImplemented", err)
	}
	if accounts != nil {
		t.Fatalf("accounts = %v, want nil", accounts)
	}
}
