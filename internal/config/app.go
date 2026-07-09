package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/go-playground/validator/v10"

	bridgev1 "car-bridge/internal/delivery/grpc/gen/bridge/v1"
	grpchandler "car-bridge/internal/delivery/grpc/handler"
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

// BootstrapConfig carries the dependencies wired by Bootstrap.
type BootstrapConfig struct {
	App      *fiber.App
	GRPC     *grpc.Server // nil to skip gRPC registration (e.g. tests)
	DB       *pgxpool.Pool
	Redis    *redis.Client // nil when Redis is not configured
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

// Bootstrap wires middleware, integrations, repositories, services, controllers
// and routes. This is the single composition root — swap a provider here.
func Bootstrap(b *BootstrapConfig) {
	// --- middleware ---
	b.App.Use(recover.New())
	b.App.Use(requestid.New())
	b.App.Use(mw.RequestLogger(b.Log)) // logger wraps the limiter so 429s are logged
	b.App.Use(newRateLimiter(b))

	// --- integrations (one base HTTP client per provider) ---
	saferwebCfg := httpConfig(b.Config, "saferweb")
	saferwebClient := saferweb.New(saferwebCfg, integrations.NewClient(saferwebCfg, b.Log))

	transgovCfg := httpConfig(b.Config, "transgov")
	transgovClient := transgov.New(transgovCfg, integrations.NewClient(transgovCfg, b.Log))

	plaidCfg := httpConfig(b.Config, "plaid")
	plaidClient := plaid.New(plaidCfg, integrations.NewClient(plaidCfg, b.Log))

	// Available for upcoming services; referenced here so wiring is explicit.
	_ = transgovClient // VINDecoder
	_ = plaidClient    // BankVerifier

	b.Log.WithFields(logrus.Fields{
		"vehicle": []string{"transgov", "saferweb"},
		"finance": []string{"plaid"},
	}).Info("integrations wired")

	// --- repositories ---
	companyRepo := repository.NewSaferwebCompanyRepository(b.DB, b.Log)
	requestLogRepo := repository.NewRequestLogRepository(b.DB, b.Log)

	// --- services (depend on ports, not vendors) ---
	carrierService := service.NewCarrierService(b.Log, saferwebClient, companyRepo)
	requestLogService := service.NewRequestLogService(b.Log, requestLogRepo)
	_ = requestLogService // API cost/outcome sink; integrations will call Record per upstream call

	// --- controllers ---
	healthController := controller.NewHealthController(b.DB, b.Redis, b.Log)
	carrierController := controller.NewCarrierController(carrierService, b.Validate, b.Log)

	// --- HTTP routes ---
	routes := route.Config{
		App:               b.App,
		HealthController:  healthController,
		CarrierController: carrierController,
	}
	routes.Setup()

	// --- gRPC services (share the same services as the REST controllers) ---
	if b.GRPC != nil {
		bridgev1.RegisterCarrierServiceServer(b.GRPC, grpchandler.NewCarrierHandler(carrierService, b.Validate, b.Log))
		bridgev1.RegisterHealthServiceServer(b.GRPC, grpchandler.NewHealthHandler(b.DB, b.Redis, b.Log))
		// Server reflection lets grpcurl/grpcui discover services without the .proto.
		reflection.Register(b.GRPC)
	}
}

// newRateLimiter builds the inbound rate limiter. It is Redis-backed when Redis
// is configured (shared limits across instances) and falls back to in-memory
// storage otherwise.
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
		Max:        max,
		Expiration: time.Duration(window) * time.Second,
		// Explicit so the keying intent is clear. c.IP() respects the proxy
		// settings configured in NewFiber (see app.proxy_header).
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

// httpConfig builds an integrations.HTTPConfig from the integrations.<name> block.
func httpConfig(v *viper.Viper, name string) integrations.HTTPConfig {
	prefix := "integrations." + name + "."
	return integrations.HTTPConfig{
		BaseURL:        v.GetString(prefix + "base_url"),
		APIKey:         v.GetString(prefix + "api_key"),
		TimeoutSeconds: v.GetInt(prefix + "timeout_seconds"),
		Retries:        v.GetInt(prefix + "retries"),
	}
}
