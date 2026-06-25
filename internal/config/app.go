package config

import (
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-playground/validator"
	"github.com/go-playground/validator/v10"

	"car-bridge/internal/integrations"
)

// Carries all dependencies
type BootstrapConfig struct {
	App      *fiber.App
	DB       *pgxpool.Pool
	Redis    *redis.Client
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(b *BootstrapConfig) {

	// Middleware
	b = BootstrapMiddleware(b)

	// Saferweb, implement later

	// Transfov, implement later

	// Plaid, implement later

	// Routes here

}

func newRateLimiter(b *BootstrapConfig) fiber.Handler {

	max := b.Config.GetInt("ratelimit.max")
	if max <= 0 {
		max = 100
	}
	window := b.Config.GetInt("ratelimit.window_seconds")
	if window <= 0 {
		window = 60
	}

	cfg := limiter.Config{
		Max:          max,
		Expiration:   time.Duration(window) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.NewError(
				fiber.StatusTooManyRequests,
				"rate limit exceeded",
			)
		},
	}

	if b.Redis != nil {
		// Implement later
	}
	return limiter.New(cfg)

}

func httpConfig(
	v *viper.Viper,
	name string,
) integrations.HTTPConfig {

	prefix := "integrations." + name + "."

	return integrations.HTTPConfig{
		BaseURL:        v.GetString(prefix + "base_url"),
		APIKey:         v.GetString(prefix + "api_key"),
		TimeoutSeconds: v.GetInt(prefix + "timeout_seconds"),
		Retries:        v.GetInt(prefix + "retries"),
	}

}

func BootstrapMiddleware(b *BootstrapConfig) *BootstrapConfig {

	b.App.User(recover.New())
	b.App.User(requestid.New())
	b.App.Use(newRateLimiter(b))

	return b

}
