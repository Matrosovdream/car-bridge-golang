package config

import (
	"errors"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func NewFiber(v *viper.Viper) *fiber.App {

	cfg := fiber.Config{
		AppName:      v.GetString("app.name"),
		ErrorHandler: errorHandler,
	}

	if header := v.GetString("app.proxy_header"); header != "" {

		cfg.ProxyHeader = header
		if proxies := v.GetStringSlice("app.trusted_proxies"); len(proxies) > 0 {
			cfg.EnableTrustedProxyCheck = true
			cfg.TrustedProxies = proxies
		}

	}

	return fiber.New(cfg)

}

func errorHandler(
	ctx *fiber.Ctx, err error,
) error {

	code := fiber.StatusInternalServerError
	var fe *fiber.Error

	if errors.As(err, &fe) {
		code = fe.code
	}

	return ctx.Status(code).JSON(
		fiber.Map{
			"errors": err.Error(),
		},
	)

}
