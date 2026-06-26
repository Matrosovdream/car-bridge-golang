package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const opTimeout = 300 * time.Millisecond

type RedisStore struct {
	client *redis.Client
	log    *logrus.Logger
}

func NewRedisStore(
	client *redis.Client, log *logrus.Logger,
) *RedisStore {

	return &RedisStore{
		client: client,
		log:    log,
	}

}

var _ fiber.Storage = (*RedisStore)(nil)

func (s *RedisStore) Get(key string) ([]byte, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		opTimeout,
	)
	defer cancel()

	val, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		s.warn("get", err)
	}

	return val, err

}

func (s *RedisStore) Set(
	key string, val []byte, exp time.Duration,
) error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		opTimeout,
	)
	defer cancel()

	if err := s.client.Set(ctx, key, val, exp).Err(); err != nil {
		s.warn("set", err)
		return err
	}

	return nil

}

func (s *RedisStore) Delete(key string) error {

	ctx, cancel := context.WithTimeout(
		context.Background(), opTimeout,
	)
	defer cancel()

	return s.client.Delete(ctx, key).Err()

}

func (s *RedisStore) Reset() error {
	return nil
}

func (s *RedisStore) Close() error {
	return nil
}

func (s *RedisStore) warn(
	op string, err error,
) {

	if s.log != nil {
		s.log.WithError(err).Warnf("rate-limiter redis %s failed; limiter fail-open", op)
	}

}
