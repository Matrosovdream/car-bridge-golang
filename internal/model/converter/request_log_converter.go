package converter

import (
	"car-bridge/internal/entity"
	"car-bridge/internal/model"
)

// defaultCurrency is applied when an entry omits one.
const defaultCurrency = "USD"

// LogEntryToEntity maps a caller's API-log entry into a persistable entity,
func LogEntryToEntity(e model.APILogEntry) *entity.RequestLog {
	currency := e.Currency
	if currency == "" {
		currency = defaultCurrency
	}
	var errMsg string
	if e.Err != nil {
		errMsg = e.Err.Error()
	}
	return &entity.RequestLog{
		ServiceType:     e.ServiceType,
		ServiceCode:     e.ServiceCode,
		Operation:       e.Operation,
		RequestRef:      e.RequestRef,
		StatusCode:      e.StatusCode,
		Success:         e.Success,
		LatencyMs:       e.LatencyMs,
		Error:           errMsg,
		Cost:            e.Cost,
		Currency:        currency,
		RequestPayload:  e.RequestPayload,
		ResponsePayload: e.ResponsePayload,
	}
}
