package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Forza Motorsport 8 dash packet format
// https://support.forzamotorsport.net/hc/en-us/articles/21742934024211-Forza-Motorsport-Data-Out-Documentation
type TelemetryPoint struct {
	OnTrack     int32  // = 1 when race is on. = 0 when in menus/race stopped
	TimestampMS uint32 // Can overflow to 0 eventually

	EngineMaxRPM     float32
	EngineIdleRPM    float32
	EngineCurrentRPM float32

	// In the car's local space; X = right, Y = up, Z = forward
	AccelarationX float32
	AccelerationY float32
	AccelerationZ float32

	// In the car's local space; X = right, Y = up, Z = forward
	VelocityX float32
	VelocityY float32
	VelocityZ float32

	// In the car's local space; X = pitch, Y = yaw, Z = roll
	AngularVelocityX float32
	AngularVelocityY float32
	AngularVelocityZ float32

	Yaw   float32
	Pitch float32
	Roll  float32

	// Suspension travel normalized: 0.0f = max stretch; 1.0 = max compression
	NormalizedSuspensionTravelFrontLeft  float32
	NormalizedSuspensionTravelFrontRight float32
	NormalizedSuspensionTravelRearLeft   float32
	NormalizedSuspensionTravelRearRight  float32

	// Tire normalized slip ratio, = 0 means 100% grip and |ratio| > 1.0 means loss of grip.
	TireSlipRatioFrontLeft  float32
	TireSlipRatioFrontRight float32
	TireSlipRatioRearLeft   float32
	TireSlipRatioRearRight  float32

	// Wheels rotation speed radians/sec.
	WheelRotationSpeedFrontLeft  float32
	WheelRotationSpeedFrontRight float32
	WheelRotationSpeedRearLeft   float32
	WheelRotationSpeedRearRight  float32

	// = 1 when wheel is on rumble strip, = 0 when off.
	WheelOnRumbleStripFrontLeft  int32
	WheelOnRumbleStripFrontRight int32
	WheelOnRumbleStripRearLeft   int32
	WheelOnRumbleStripRearRight  int32

	// = from 0 to 1, where 1 is the deepest puddle
	WheelInPuddleDepthFrontLeft  float32
	WheelInPuddleDepthFrontRight float32
	WheelInPuddleDepthRearLeft   float32
	WheelInPuddleDepthRearRight  float32

	// Non-dimensional surface rumble values passed to controller force feedback
	SurfaceRumbleFrontLeft  float32
	SurfaceRumbleFrontRight float32
	SurfaceRumbleRearLeft   float32
	SurfaceRumbleRearRight  float32

	// Tire normalized slip angle, = 0 means 100% grip and |angle| > 1.0 means loss of grip.
	TireSlipAngleFrontLeft  float32
	TireSlipAngleFrontRight float32
	TireSlipAngleRearLeft   float32
	TireSlipAngleRearRight  float32

	// Tire normalized combined slip, = 0 means 100% grip and |slip| > 1.0 means loss of grip.
	TireCombinedSlipFrontLeft  float32
	TireCombinedSlipFrontRight float32
	TireCombinedSlipRearLeft   float32
	TireCombinedSlipRearRight  float32

	// Actual suspension travel in meters
	SuspensionTravelMetersFrontLeft  float32
	SuspensionTravelMetersFrontRight float32
	SuspensionTravelMetersRearLeft   float32
	SuspensionTravelMetersRearRight  float32

	CarOrdinal          int32 // Unique ID of the car make/model
	CarClass            int32 // Between 0 (D -- worst cars) and 7 (X class -- best cars) inclusive
	CarPerformanceIndex int32 // Between 100 (worst car) and 999 (best car) inclusive
	DrivetrainType      int32 // 0 = FWD, 1 = RWD, 2 = AWD
	NumCylinders        int32 // Number of cylinders in the engine

	PositionX float32
	PositionY float32
	PositionZ float32

	Speed  float32 // meters per second
	Power  float32 // watts
	Torque float32 //newton meter

	// Rear right tire is wrong (same as rear left)
	// https://forums.forza.net/t/data-out-udp-incorrect-data-wheelinpuddledepthfrontleft-tiretemprearright-1932122/730417
	TireTempFrontLeft  float32
	TireTempFrontRight float32
	TireTempRearLeft   float32
	TireTempRearRight  float32

	Boost            float32
	Fuel             float32
	DistanceTraveled float32
	BestLap          float32
	LastLap          float32
	CurrentLap       float32
	CurrentRaceTime  float32

	LapNumber uint16 `bun:"type:INTEGER"`

	RacePosition uint8
	Accel        uint8
	Brake        uint8
	Clutch       uint8
	HandBrake    uint8
	Gear         uint8

	Steer                       int8
	NormalizedDrivingLine       int8
	NormalizedAIBrakeDifference int8

	TireWearFrontLeft  float32
	TireWearFrontRight float32
	TireWearRearLeft   float32
	TireWearRearRight  float32

	TrackOrdinal int32 // Track ID
}

type Point struct {
	bun.BaseModel
	TelemetryPoint

	Race      uuid.UUID `bun:"type:uuid"`
	CreatedAt time.Time
}

func (p Point) ToProto() *ApiPoint {
	return &ApiPoint{
		RaceTime:           p.CurrentRaceTime,
		LapTime:            p.CurrentLap,
		LapNumber:          uint32(p.LapNumber),
		Fuel:               p.Fuel,
		Speed:              p.Speed,
		RacePosition:       uint32(p.RacePosition),
		Accel:              uint32(p.Accel),
		Brake:              uint32(p.Brake),
		Gear:               uint32(p.Gear),
		TireWearFrontLeft:  p.TireWearFrontLeft,
		TireWearFrontRight: p.TireWearFrontRight,
		TireWearRearLeft:   p.TireWearRearLeft,
		TireWearRearRight:  p.TireWearRearRight,
		TireTempFrontLeft:  p.TireTempFrontLeft,
		TireTempFrontRight: p.TireTempFrontRight,
		TireTempRearLeft:   p.TireTempRearLeft,
		TireTempRearRight:  p.TireTempRearRight,
		PositionX:          p.PositionX,
		PositionY:          p.PositionY,
		PositionZ:          p.PositionZ,
		EngineCurrentRPM:   p.EngineCurrentRPM,
		Steer:              int32(p.Steer),
	}
}
