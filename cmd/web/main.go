package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

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
	grpcServer := config.NewGRPCServer(v, log)

	// One composition root registers both the REST routes and the gRPC services,
	// so they share the same underlying service instances.
	config.Bootstrap(&config.BootstrapConfig{
		App:      app,
		GRPC:     grpcServer,
		DB:       db,
		Redis:    rdb,
		Log:      log,
		Validate: validate,
		Config:   v,
	})

	// --- gRPC server (separate port) ---
	grpcPort := v.GetInt("grpc.port")
	if grpcPort == 0 {
		grpcPort = 50051
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen on %s: %v", grpcAddr, err)
	}
	go func() {
		log.Infof("car-bridge gRPC listening on %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	// --- HTTP server (Fiber) ---
	port := v.GetInt("app.port")
	if port == 0 {
		port = 3000
	}
	httpAddr := fmt.Sprintf(":%d", port)
	go func() {
		log.Infof("car-bridge HTTP listening on %s", httpAddr)
		if err := app.Listen(httpAddr); err != nil {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// --- graceful shutdown on SIGINT/SIGTERM ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down…")
	grpcServer.GracefulStop()
	if err := app.Shutdown(); err != nil {
		log.Errorf("http shutdown error: %v", err)
	}
}
