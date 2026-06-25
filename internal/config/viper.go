package config

import (
	"strings"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {

	// Set Config
	v := PrepareViperConfig()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(err)
		}
	}

	return v

}

func PrepareViperConfig() *viper.Viper {

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./../../")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("redis.url", "REDIS_URL")
	_ = v.BindEnv("integrations.saferweb.api_key", "SAFERWEB_WEBKEY")
	_ = v.BindEnv("integrations.plaid.api_key", "PLAID_SECRET")

	return v

}
