package config

import (
	"strings"

	"github.com/spf13/viper"
)

// NewViper loads config.yaml and overlays environment variables. Secrets are
// bound explicitly so DATABASE_URL / *_WEBKEY etc. override the yaml defaults.
func NewViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./../../")

	// Map nested keys to underscore env vars (app.port -> APP_PORT,
	// ratelimit.max -> RATELIMIT_MAX) and let an explicit empty env value win
	// (so REDIS_URL="" selects the in-memory limiter).
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("integrations.saferweb.api_key", "SAFERWEB_WEBKEY")
	_ = v.BindEnv("integrations.carsxe.api_key", "CARSXE_API_KEY")
	_ = v.BindEnv("integrations.vehicledatabases.api_key", "VEHICLEDATABASES_API_KEY")
	_ = v.BindEnv("integrations.plaid.api_key", "PLAID_SECRET")
	_ = v.BindEnv("integrations.postmark.api_key", "POSTMARK_SERVER_TOKEN")
	_ = v.BindEnv("integrations.sendgrid.api_key", "SENDGRID_API_KEY")
	_ = v.BindEnv("integrations.twilio.api_key", "TWILIO_AUTH_TOKEN")

	if err := v.ReadInConfig(); err != nil {
		// A missing file is fine (env-only deploys); anything else is fatal.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}
	return v
}
