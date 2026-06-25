package config

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRedis(
	ctx context.Context,
	v *viper.Viper,
	log *logrus.Logger,
) *redis.Client {

	url := v.GetString("redis.url")
	if url == "" {
		log.Warn("redis.url not set; rate limiter will use in-memory storage")
		return nil
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Errorf("invalid redis url %q; falling back to in-memory rate limiter: %v", url, err)
		return nil
	}

	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		log.Warnf("redis not reachable at startup: %v", err)
	}

	return client

}
