package handler

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	bridgev1 "car-bridge/internal/delivery/grpc/gen/bridge/v1"
)

// HealthHandler implements bridgev1.HealthServiceServer. It mirrors
// controller.HealthController.Check: same dependency probes, same status
// vocabulary. gRPC has no HTTP status codes, so a degraded state is reported in
// the CheckResponse body (status="degraded") rather than as an RPC error, which
// keeps the dependency breakdown available to the caller.
type HealthHandler struct {
	bridgev1.UnimplementedHealthServiceServer
	DB    *pgxpool.Pool
	Redis *redis.Client // nil when Redis is not configured
	Log   *logrus.Logger
}

// NewHealthHandler constructs the health gRPC handler.
func NewHealthHandler(db *pgxpool.Pool, rdb *redis.Client, log *logrus.Logger) *HealthHandler {
	return &HealthHandler{DB: db, Redis: rdb, Log: log}
}

// Check reports service + dependency health.
func (h *HealthHandler) Check(ctx context.Context, _ *bridgev1.CheckRequest) (*bridgev1.CheckResponse, error) {
	c, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if h.DB == nil || h.DB.Ping(c) != nil {
		dbStatus = "down"
	}

	redisStatus := "disabled"
	if h.Redis != nil {
		if h.Redis.Ping(c).Err() != nil {
			redisStatus = "down"
		} else {
			redisStatus = "ok"
		}
	}

	overall := "ok"
	if dbStatus == "down" || redisStatus == "down" {
		overall = "degraded"
	}

	return &bridgev1.CheckResponse{
		Status: overall,
		Db:     dbStatus,
		Redis:  redisStatus,
	}, nil
}
