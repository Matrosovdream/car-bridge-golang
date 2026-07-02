package model

// APILogEntry is the input a caller supplies to record one upstream API call.
type APILogEntry struct {
	ServiceType     string
	ServiceCode     string
	Operation       string
	RequestRef      string
	StatusCode      *int
	Success         bool
	LatencyMs       *int
	Err             error
	Cost            float64
	Currency        string
	RequestPayload  []byte
	ResponsePayload []byte
}
