// Package finance defines payments & identity capability ports.
// Providers (plaid, kyc, apr) implement these.
package finance

import "context"

// BankAccount is a normalized account returned by a bank-verification provider.
type BankAccount struct {
	AccountID string
	Name      string
	Type      string
	Balance   float64
	Currency  string
}

// IdentityCheck is the result of a KYC/identity verification.
type IdentityCheck struct {
	Subject string
	Passed  bool
	Reasons []string
}

// LoanQuote is a computed financing quote.
type LoanQuote struct {
	Principal  float64
	APR        float64
	TermMonths int
	Monthly    float64
}

// BankVerifier reads bank/income data (e.g. Plaid).
type BankVerifier interface {
	Accounts(ctx context.Context, accessToken string) ([]BankAccount, error)
}

// IdentityVerifier runs a KYC/credit check.
type IdentityVerifier interface {
	Verify(ctx context.Context, subject string) (*IdentityCheck, error)
}

// LoanCalculator computes monthly payment / APR figures.
type LoanCalculator interface {
	Quote(ctx context.Context, principal, apr float64, termMonths int) (*LoanQuote, error)
}
