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

	"car-bridge/internal/delivery/http/controller"
	mw "car-bridge/internal/delivery/http/middleware"
	"car-bridge/internal/delivery/http/route"
	"car-bridge/internal/integrations"
	"car-bridge/internal/integrations/finance/plaid"
	"car-bridge/internal/integrations/vehicle/saferweb"
	"car-bridge/internal/integrations/vehicle/transgov"
	"car-bridge/internal/repository"
	"car-bridge/internal/service"
)

type BootstrapConfig struct {
	App      *fiber.App
	DB       *pgxpool.Pool
	Redis    *redis.Client
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(b *BootstrapConfig) {

	b.App.Use(recover.New())
	b.App.Use(requestid.New())
	b.App.Use(mw.RequestLogger(b.Log))
	b.App.Use(newRateLimiter(b))

	// Integrations
	saferwebCfg := httpConfig(b.Config, "saferweb")
	saferwebClient := saferweb.New(
		saferwebCfg,
		integrations.NewClient(saferwebCfg, b.Log),
	)

	transgovCfg := httpConfig(b.Config, "transgov")
	transgovClient := transgov.New(
		transgovCfg,
		integrations.NewClient(transgovCfg, b.Log),
	)

	plaidCfg := httpConfig(b.Config, "plaid")
	plaidClient := plaid.New(
		plaidCfg,
		integrations.NewClient(plaidCfg, b.Log),
	)

	_ = transgovClient
	_ = plaidClient

	b.Log.WithFields(
		logrus.Fields{
			"vehicle": []string{"transgov", "saferweb"},
			"finance": []string{"plaid"},
		},
	).Info("Integrations wired")

	companyRepo := repository.NewSaferwebCompanyRepository(
		d.DB, b.Log,
	)

	carrierService := service.NewCarrierService(
		b.Log, saferwebClient, companyRepo,
	)

	healthController := controller.NewHealthController(
		b.DB, b.Redis, b.Log,
	)
	carrierController := controller.NewCarrierController(
		carrierService, b.Validate, b.Log,
	)

	routes := route.Config{
		App:               b.App,
		HealthController:  healthController,
		CarrierController: carrierController,
	}
	routes.Setup()

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
			return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded")
		},
	}

	if b.Redis != nil {
		cfg.Storage = mw.NewRedisStore(b.Redis, b.Log)
	}

	return limiter.New(cfg)

}

func httpConfig(
	v *viper.Viper, name string,
) integrations.HTTPConfig {

	prefix := "integrations." + name + "."

	return integrations.HTTPConfig{
		BaseURL:        v.GetString(prefix + "base_url"),
		APIKey:         v.GetString(prefix + "api_key"),
		TimeoutSeconds: v.GetInt(prefix + "timeout_seconds"),
		Retries:        v.GetInt(prefix + "retries"),
	}

}
