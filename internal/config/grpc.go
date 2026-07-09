package config

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewGRPCServer builds a bare *grpc.Server with recovery + logging interceptors.
// Services are registered later, in Bootstrap, so gRPC handlers share the exact
// same service instances as the REST controllers.
//
// The viper handle is accepted for symmetry with NewFiber and to leave room for
// server options (keepalive, TLS, message sizes) sourced from config.
func NewGRPCServer(_ *viper.Viper, log *logrus.Logger) *grpc.Server {
	return grpc.NewServer(
		// recover first (outermost) so it also catches panics from logging.
		grpc.ChainUnaryInterceptor(
			recoveryUnaryInterceptor(log),
			loggingUnaryInterceptor(log),
		),
	)
}

// loggingUnaryInterceptor logs one line per unary call: method, resolved status
// code, and latency. It is the gRPC analogue of the HTTP RequestLogger.
func loggingUnaryInterceptor(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		entry := log.WithFields(logrus.Fields{
			"method":  info.FullMethod,
			"code":    status.Code(err).String(),
			"latency": time.Since(start).String(),
		})
		if err != nil {
			entry.WithError(err).Warn("grpc call")
		} else {
			entry.Info("grpc call")
		}
		return resp, err
	}
}

// recoveryUnaryInterceptor converts a panicking handler into an INTERNAL error
// instead of crashing the server, mirroring the HTTP recover middleware.
func recoveryUnaryInterceptor(log *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(logrus.Fields{
					"method": info.FullMethod,
					"panic":  r,
				}).Error("grpc handler panicked")
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}
