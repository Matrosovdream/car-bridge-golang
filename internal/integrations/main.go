package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"

	"car-bridge/internal/config"
)

func main() {
	// Load .env for local dev (ignored if absent).
	_ = godotenv.Load()

	v := config.NewViper()
	log := config.NewLogger(v)
	validate := config.NewValidator()

	ctx := context.Background()
	db := config.NewDatabase(ctx, v, log)
	defer db.Close()

	rdb := config.NewRedis(ctx, v, log)
	if rdb != nil {
		defer rdb.Close()
	}

	app := config.NewFiber(v)
	config.Bootstrap(&config.BootstrapConfig{
		App:      app,
		DB:       db,
		Redis:    rdb,
		Log:      log,
		Validate: validate,
		Config:   v,
	})

	port := v.GetInt("app.port")
	if port == 0 {
		port = 3000
	}
	addr := fmt.Sprintf(":%d", port)
	log.Infof("car-bridge listening on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
