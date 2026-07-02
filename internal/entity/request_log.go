package entity

import "time"

// RequestLog is the History table for all API calls
type RequestLog struct {
	ID              int64
	ServiceType     string
	ServiceCode     string
	Operation       string
	RequestRef      string
	StatusCode      *int // HTTP status; nil when no response was received
	Success         bool
	LatencyMs       *int    // round-trip latency; nil when unknown
	Error           string  // failure message; "" ⇒ NULL
	Cost            float64 // NUMERIC(12,4) — provider charge for this call
	Currency        string  // ISO 4217, defaults to USD
	RequestPayload  []byte  // JSONB, nullable
	ResponsePayload []byte  // JSONB, nullable
	CreatedAt       time.Time
}
