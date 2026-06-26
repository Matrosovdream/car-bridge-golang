package geo

import "context"

type Place struct {
	Query     string
	Latitude  float64
	Longitude float64
	Formatted string
}

type Matrix struct {
	Origins      []string
	Destinations []string
	Distances    [][]float64
	Durations    [][]float64
}

type FuelStation struct {
	ID        string
	Name      string
	FuelType  string
	Latitude  float64
	Longitude float64
}

type Geocoder interface {
	Geocode(ctx context.Context, query string) (*Place, error)
}

type RouteMatrix interface {
	DistanceMatrix(ctx context.Context, origins, destinations []string) (*Matrix, error)
}

type FuelStationFinder interface {
	NearestStations(ctx context.Context, lat, lng float64, limit int) ([]FuelStation, error)
}
