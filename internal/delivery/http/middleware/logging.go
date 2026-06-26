package middleware

import (
	"errors"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func RequestLogger(log *logrus.Logger) fiber.Handler {

	return func(c *fiber.Ctx) error {

		start := time.Now()
		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			status = fiber.StatusInternalServerError

			var fe *fiber.Error

			if errors.As(err, &fe) {
				status = fe.Code
			}
		}

		log.WithFields(
			logrus.Fields{
				"method":     c.Method(),
				"path":       c.Path(),
				"status":     status,
				"latency_ms": time.Since(start).Milliseconds(),
				"request_id": c.Locals("requestid"),
			},
		).Info("request")

		return err

	}

}
