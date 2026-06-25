package config

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewDatabase(
	ctx context.Context,
	v *viper.Viper,
	log *logrus.Logger,
) *pgxpool.Pool {

	dsn := v.GetString("database.url")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("invalid database config: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Warnf("database not reachable at startup: %v", err)
	}

	return pool

}
