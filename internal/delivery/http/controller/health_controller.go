package controller

import (
	"context"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type HealthController struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Log   *logrus.Logger
}

func NewHealthController(
	db *pgxpool.Pool,
	rdb *redis.Client,
	log *logrus.Logger,
) *HealthController {

	return &HealthController{
		DB:    db,
		Redis: rdb,
		Log:   log,
	}

}

func (h *HealthController) Check(ctx *fiber.Ctx) error {

	c, cancel := context.WithTimeout(
		ctx.UserContext(),
		2*time.Second,
	)
	defer cancel()

	dbStatus := "ok"
	if h.DB == nil || h.DB.Ping(c) != nil {
		dbStatus = "down"
	}

	redisStatus := "disable"
	if h.Redis != nil {

		if h.Redis.Ping(c).Err() != nil {
			redisStatus = "down"
		} else {
			redisStatus = "ok"
		}

	}

	overall, code := "ok", fiber.StatusOK
	if dbStatus == "down" || redisStatus == "down" {
		overall, code = "degraded", fiber.StatusServiceUnavailable
	}

	return ctx.Status(code).JSON(
		fiber.Map{
			"status": overall,
			"db":     dbStatus,
			"redis":  redisStatus,
		},
	)

}
