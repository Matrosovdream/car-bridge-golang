// Package apr will implement finance.LoanCalculator. This one is likely a local
// computation (amortization) rather than an external API, but lives here so the
// finance service depends only on the port.
//
// TODO: replace this placeholder with client.go:
//   - New(...) *Client
//   - Quote(ctx, principal, apr, termMonths) (*finance.LoanQuote, error)
package apr
