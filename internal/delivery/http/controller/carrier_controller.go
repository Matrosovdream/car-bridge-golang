package controller

import (
	"errors"

	"github.com/go-playground/validator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"

	"car-bridge/internal/integrations"
	"car-bridge/internal/model"
	"car-bridge/internal/service"
)

type CarrierController struct {
	Service  *service.CarrierService
	Validate *validator.Validate
	Log      *logrus.Logger
}

func NewCarrierController(
	svc *service.CarrierService,
	validate *validator.Validate,
	log *logrus.Logger,
) *CarrierController {

	return &CarrierController{
		Service:  svc,
		Validate: validate,
		Log:      log,
	}

}

func (c *CarrierController) GetByDOT(
	ctx *fiber.Ctx,
) error {

	req := &model.CarrierRequest{
		DOTNumber: ctx.Params("dot"),
	}
	if err := c.Validate.Struct(req); err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest, "invalid dot number",
		)
	}

	res, err := c.Service.GetByDOT(
		ctx.UserContext(), req.DOTNumber,
	)
	if err != nil {
		return c.mapError()
	}

	return ctx.JSON(
		model.WebResponse[*model.CarrierResponse]{
			Data: res,
		},
	)

}

func (c *CarrierController) mapError(err error) error {

	switch {
	case errors.Is(err, integrations.ErrNotImplemented):
		return fiber.NewError(fiber.StatusNotImplemented, "carrier lookup not implemented yet")
	case errors.Is(err, integrations.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "carrier not found")
	case errors.Is(err, integrations.ErrUpstreamUnavailable):
		return fiber.NewError(fiber.StatusBadGateway, "carrier data source unavailable")
	case errors.Is(err, integrations.ErrRateLimited):
		return fiber.NewError(fiber.StatusTooManyRequests, "carrier data source rate limited")
	default:
		c.Log.WithError(err).Error("carrier lookup failed")
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}

}
