// Package telematics defines the connected-car capability port.
// Providers (smartcar) implement it.
package telematics

import "context"

// VehicleState is a normalized snapshot read from a connected vehicle.
type VehicleState struct {
	VehicleID      string
	OdometerKM     float64
	Latitude       float64
	Longitude      float64
	FuelPercent    *float64 // nil if unsupported
	BatteryPercent *float64 // nil if not an EV
	Locked         *bool
}

// ConnectedVehicle reads state from and sends commands to a real vehicle.
type ConnectedVehicle interface {
	State(ctx context.Context, vehicleID string) (*VehicleState, error)
	Lock(ctx context.Context, vehicleID string) error
	Unlock(ctx context.Context, vehicleID string) error
}
