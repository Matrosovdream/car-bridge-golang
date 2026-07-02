package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"car-bridge/internal/entity"
)

// RequestLogRepository appends upstream API-call records for audit/cost tracking.
type RequestLogRepository struct {
	DB  *pgxpool.Pool
	Log *logrus.Logger
}

// NewRequestLogRepository constructs the repository.
func NewRequestLogRepository(db *pgxpool.Pool, log *logrus.Logger) *RequestLogRepository {
	return &RequestLogRepository{DB: db, Log: log}
}

// Insert appends one log row, filling ID and CreatedAt from the DB. Empty
// request_ref/error are stored as NULL.
func (r *RequestLogRepository) Insert(ctx context.Context, e *entity.RequestLog) error {
	const q = `
INSERT INTO request_log
	(service_type, service_code, operation, request_ref, status_code,
	 success, latency_ms, error, cost, currency, request_payload, response_payload)
VALUES
	($1, $2, $3, NULLIF($4, ''), $5, $6, $7, NULLIF($8, ''), $9, $10, $11, $12)
RETURNING id, created_at`

	return r.DB.QueryRow(ctx, q,
		e.ServiceType, e.ServiceCode, e.Operation, e.RequestRef, e.StatusCode,
		e.Success, e.LatencyMs, e.Error, e.Cost, e.Currency, e.RequestPayload, e.ResponsePayload,
	).Scan(&e.ID, &e.CreatedAt)
}
