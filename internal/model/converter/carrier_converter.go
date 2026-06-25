package converter

import (
	"car-bridge/internal/entity"
	"car-bridge/internal/integrations/vehicle"
	"car-bridge/internal/model"
)

func CarrierToEntity(c *vehicle.Carrier) *entity.SaferwebCompany {

	if c == nil {
		return nil
	}

	return &entity.SaferwebCompany{
		DOTNumber: c.DOTNumber,
		LegalName: c.LegalName,
		DBAName:   c.DBAName,
	}

}

func CarrierToResponse(e *entity.SaferwebCompany) *model.CarrierResponse {

	if e == nil {
		return nil
	}
	return &model.CarrierResponse{
		DOTNumber: e.DOTNumber,
		LegalName: e.LegalName,
		DBAName:   e.DBAName,
	}

}
