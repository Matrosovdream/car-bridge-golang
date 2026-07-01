// Package geo defines location & routing capability ports.
// Providers (mapbox, googlemaps, nrel) implement these.
package geo

import "context"

// Place is a geocoded location.
type Place struct {
	Query     string
	Latitude  float64
	Longitude float64
	Formatted string
}

// Matrix holds pairwise distances/durations between origins and destinations.
type Matrix struct {
	Origins      []string
	Destinations []string
	Distances    [][]float64 // meters
	Durations    [][]float64 // seconds
}

// FuelStation is an alternative-fuel / charging location (NREL).
type FuelStation struct {
	ID        string
	Name      string
	FuelType  string
	Latitude  float64
	Longitude float64
}

// Geocoder turns an address/query into coordinates.
type Geocoder interface {
	Geocode(ctx context.Context, query string) (*Place, error)
}

// RouteMatrix computes pairwise distances/durations.
type RouteMatrix interface {
	DistanceMatrix(ctx context.Context, origins, destinations []string) (*Matrix, error)
}

// FuelStationFinder locates nearby fuel/charging stations.
type FuelStationFinder interface {
	NearestStations(ctx context.Context, lat, lng float64, limit int) ([]FuelStation, error)
}
