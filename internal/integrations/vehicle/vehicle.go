// Package vehicle defines capability ports for vehicle & motor-carrier data.
// Providers (transgov, saferweb, carsxe, vehicledatabases) implement these.
package vehicle

import "context"

// Carrier is the normalized motor-carrier record (FMCSA SAFER / QCMobile).
type Carrier struct {
	DOTNumber    string
	LegalName    string
	DBAName      string
	EntityType   string
	PhyState     string
	Status       string
	TotalTrucks  int
	TotalDrivers int
}

// VehicleSpec is the normalized result of decoding a VIN (NHTSA vPIC, etc.).
type VehicleSpec struct {
	VIN       string
	Make      string
	Model     string
	ModelYear int
	BodyClass string
	FuelType  string
}

// CarrierLookup resolves carrier safety/registration data by DOT number.
type CarrierLookup interface {
	LookupByDOT(ctx context.Context, dotNumber string) (*Carrier, error)
}

// VINDecoder turns a VIN into vehicle specifications.
type VINDecoder interface {
	Decode(ctx context.Context, vin string) (*VehicleSpec, error)
}

// PlateDecoder resolves a license plate (+ state) to a vehicle.
type PlateDecoder interface {
	DecodePlate(ctx context.Context, plate, state string) (*VehicleSpec, error)
}
