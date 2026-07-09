// Package handler holds the gRPC service implementations. Each handler is the
// gRPC counterpart of an HTTP controller and delegates to the same service
// layer, so REST and gRPC always share one code path and one set of behaviours.
package handler

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bridgev1 "car-bridge/internal/delivery/grpc/gen/bridge/v1"
	"car-bridge/internal/integrations"
	"car-bridge/internal/model"
	"car-bridge/internal/service"
)

// CarrierHandler implements bridgev1.CarrierServiceServer. It mirrors
// controller.CarrierController: same service, same validator, same error map.
type CarrierHandler struct {
	bridgev1.UnimplementedCarrierServiceServer
	Service  *service.CarrierService
	Validate *validator.Validate
	Log      *logrus.Logger
}

// NewCarrierHandler constructs the carrier gRPC handler.
func NewCarrierHandler(svc *service.CarrierService, validate *validator.Validate, log *logrus.Logger) *CarrierHandler {
	return &CarrierHandler{Service: svc, Validate: validate, Log: log}
}

// GetByDOT looks up a motor carrier by its USDOT number.
func (h *CarrierHandler) GetByDOT(ctx context.Context, in *bridgev1.GetByDOTRequest) (*bridgev1.GetByDOTResponse, error) {
	// Reuse the REST request model so validation rules stay in one place.
	req := &model.CarrierRequest{DOTNumber: in.GetDotNumber()}
	if err := h.Validate.Struct(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid dot number")
	}

	res, err := h.Service.GetByDOT(ctx, req.DOTNumber)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &bridgev1.GetByDOTResponse{
		Carrier: &bridgev1.Carrier{
			DotNumber: res.DOTNumber,
			LegalName: res.LegalName,
			DbaName:   res.DBAName,
		},
	}, nil
}

// mapError translates integration errors into gRPC status codes, the same way
// CarrierController.mapError translates them into HTTP status codes.
func (h *CarrierHandler) mapError(err error) error {
	switch {
	case errors.Is(err, integrations.ErrNotImplemented):
		return status.Error(codes.Unimplemented, "carrier lookup not implemented yet")
	case errors.Is(err, integrations.ErrNotFound):
		return status.Error(codes.NotFound, "carrier not found")
	case errors.Is(err, integrations.ErrUpstreamUnavailable):
		return status.Error(codes.Unavailable, "carrier data source unavailable")
	case errors.Is(err, integrations.ErrRateLimited):
		return status.Error(codes.ResourceExhausted, "carrier data source rate limited")
	default:
		h.Log.WithError(err).Error("carrier lookup failed")
		return status.Error(codes.Internal, "internal error")
	}
}
