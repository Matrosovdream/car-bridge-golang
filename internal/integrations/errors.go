package integrations

import "errors"

var (
	// ErrNotImplemented is returned by provider stubs that are scaffolded but not wired yet.
	ErrNotImplemented = errors.New("integration: not implemented")
	// ErrUpstreamUnavailable signals a network error or 5xx — retryable.
	ErrUpstreamUnavailable = errors.New("integration: upstream unavailable")
	// ErrNotFound maps to an upstream 404.
	ErrNotFound = errors.New("integration: resource not found")
	// ErrRateLimited maps to an upstream 429 — retryable.
	ErrRateLimited = errors.New("integration: rate limited")
	// ErrBadResponse signals a 2xx whose body could not be decoded.
	ErrBadResponse = errors.New("integration: bad response from upstream")
)
