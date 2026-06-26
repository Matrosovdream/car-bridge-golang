package telematics

import "context"

type VehicleState struct {
	VehicleID      string
	OdometerKM     float64
	Latitude       float64
	Longitude      float64
	FuelPercent    *float64
	BatteryPercent *float64
	Locked         *bool
}

type ConnectedVehicle interface {
	State(ctx context.Context, vehicleID string) (*VehicleState, error)
	Lock(ctx context.Context, vehicleID string) error
	Unlock(ctx context.Context, vehicleID string) error
}
