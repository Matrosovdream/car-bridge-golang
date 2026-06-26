package vehicle

import "context"

type Carrier struct {
	DOTNumber    string
	LegalName    string
	DBName       string
	EntityType   string
	PhyState     string
	Status       string
	TotalTrucks  int
	TotalDrivers int
}

type VehicleSpec struct {
	VIN       string
	Make      string
	Model     string
	ModelYear int
	BodyClass string
	FuelType  string
}

type CarrierLookup interface {
	LookupByDOT(ctx context.Context, dotNuber string) (*Carrier, error)
}

type VINDecoder interface {
	Decode(ctx context.Context, vin string) (*VehicleSpec, error)
}

type PlateDecoder interface {
	DecodePlate(ctx context.Context, plate, state string) (*VehicleSpec, error)
}
