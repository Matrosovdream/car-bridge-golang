package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"car-bridge/internal/integrations/vehicle"
	"car-bridge/internal/model"
	"car-bridge/internal/model/converter"
	"car-bridge/internal/repository"
)

type CarrierService struct {
	Log      *logrus.Logger
	Carriers vehicle.CarrierLookup
	Repo     *repository.SaferwebCompanyRepository
}

func NewCarrierService(
	log *logrus.Logger,
	carriers vehicle.CarrierLookup,
	repo *repository.SaferwebCompanyRepository,
) *CarrierService {

	return &CarrierService{
		Log:      log,
		Carriers: carriers,
		Repo:     repo,
	}

}

func (s *CarrierService) GetByDOT(
	ctx context.Context, dotNumber string,
) (*model.CarrierResponse, error) {

	if cached, err := s.Repo.FindByDOT(ctx, dotNumber); err != nil {
		s.Log.WithError(err).Warn("carrier cache lookup failed; falling through to upstream")
	} else if cached != nil {
		return converter.CarrierToResponse(cached), nil
	}

	carrier, err := s.Carriers.LookupByDOT(ctx, dotNumber)
	if err != nil {
		return nil, err
	}

	ent := converter.CarrierToEntity(carrier)
	if err := s.Repo.Upsert(ctx, ent); err != nil {
		s.Log.WithError(err).Warn("failed to cache carrier")
	}

	return converter.CarrierToResponse(ent), nil

}
