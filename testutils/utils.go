package testutils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"forzatelemetry/models"
	"forzatelemetry/storage"
)

func ExecuteRequest(req *http.Request, router chi.Router) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func NewStore(fixtures ...string) *storage.Store {
	store, err := storage.NewSqliteStore("file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	err = store.CreateTables(ctx)
	if err != nil {
		panic(err)
	}

	err = store.LoadFixtures(ctx, fixtures...)
	if err != nil {
		panic(err)
	}
	return store
}

type errorResponsePayload struct {
	Error   string         `json:"error"`
	Details map[string]any `json:"details"`
}

func CheckErrorPayload(resp *httptest.ResponseRecorder, errorMsg string, t *testing.T) {
	var respError errorResponsePayload
	err := json.NewDecoder(resp.Body).Decode(&respError)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if respError.Error != errorMsg {
		t.Errorf("expected %v got %v", errorMsg, respError.Error)
	}
}

func ParseTime(t string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, t)
	if err != nil {
		panic(err)
	}
	return parsed
}

func ParseUUID(id string) uuid.UUID {
	uid, err := uuid.Parse(id)
	if err != nil {
		panic(err)
	}
	return uid
}

func Point(race uuid.UUID, createdAt time.Time, onTrack int32) models.Point {
	return models.Point{
		TelemetryPoint: models.TelemetryPoint{
			OnTrack:                              onTrack,
			TimestampMS:                          1000,
			EngineMaxRPM:                         10000,
			EngineIdleRPM:                        1000,
			EngineCurrentRPM:                     5000,
			AccelarationX:                        1.100,
			AccelerationY:                        1.100,
			AccelerationZ:                        1.100,
			VelocityX:                            1.100,
			VelocityY:                            1.100,
			VelocityZ:                            1.100,
			AngularVelocityX:                     1.100,
			AngularVelocityY:                     1.100,
			AngularVelocityZ:                     1.100,
			Yaw:                                  1.100,
			Pitch:                                1.100,
			Roll:                                 1.100,
			NormalizedSuspensionTravelFrontLeft:  1.100,
			NormalizedSuspensionTravelFrontRight: 1.100,
			NormalizedSuspensionTravelRearLeft:   1.100,
			NormalizedSuspensionTravelRearRight:  1.100,
			TireSlipRatioFrontLeft:               1.100,
			TireSlipRatioFrontRight:              1.100,
			TireSlipRatioRearLeft:                1.100,
			TireSlipRatioRearRight:               1.100,
			WheelRotationSpeedFrontLeft:          1.100,
			WheelRotationSpeedFrontRight:         1.100,
			WheelRotationSpeedRearLeft:           1.100,
			WheelRotationSpeedRearRight:          1.100,
			WheelOnRumbleStripFrontLeft:          1,
			WheelOnRumbleStripFrontRight:         1,
			WheelOnRumbleStripRearLeft:           1,
			WheelOnRumbleStripRearRight:          1,
			WheelInPuddleDepthFrontLeft:          1.100,
			WheelInPuddleDepthFrontRight:         1.100,
			WheelInPuddleDepthRearLeft:           1.100,
			WheelInPuddleDepthRearRight:          1.100,
			SurfaceRumbleFrontLeft:               1.100,
			SurfaceRumbleFrontRight:              1.100,
			SurfaceRumbleRearLeft:                1.100,
			SurfaceRumbleRearRight:               1.100,
			TireSlipAngleFrontLeft:               1.100,
			TireSlipAngleFrontRight:              1.100,
			TireSlipAngleRearLeft:                1.100,
			TireSlipAngleRearRight:               1.100,
			TireCombinedSlipFrontLeft:            1.100,
			TireCombinedSlipFrontRight:           1.100,
			TireCombinedSlipRearLeft:             1.100,
			TireCombinedSlipRearRight:            1.100,
			SuspensionTravelMetersFrontLeft:      1.100,
			SuspensionTravelMetersFrontRight:     1.100,
			SuspensionTravelMetersRearLeft:       1.100,
			SuspensionTravelMetersRearRight:      1.100,
			CarOrdinal:                           1,
			CarClass:                             1,
			CarPerformanceIndex:                  100,
			DrivetrainType:                       0,
			NumCylinders:                         6,
			PositionX:                            1.100,
			PositionY:                            1.100,
			PositionZ:                            1.100,
			Speed:                                1.100,
			Power:                                1.100,
			Torque:                               1.100,
			TireTempFrontLeft:                    1.100,
			TireTempFrontRight:                   1.100,
			TireTempRearLeft:                     1.100,
			TireTempRearRight:                    1.100,
			Boost:                                1.100,
			Fuel:                                 1.100,
			DistanceTraveled:                     1.100,
			BestLap:                              1.100,
			LastLap:                              1.100,
			CurrentLap:                           1.100,
			CurrentRaceTime:                      1.100,
			LapNumber:                            2,
			RacePosition:                         2,
			Accel:                                2,
			Brake:                                2,
			Clutch:                               2,
			HandBrake:                            2,
			Gear:                                 2,
			Steer:                                2,
			NormalizedDrivingLine:                2,
			NormalizedAIBrakeDifference:          2,
			TireWearFrontLeft:                    1.100,
			TireWearFrontRight:                   1.100,
			TireWearRearLeft:                     1.100,
			TireWearRearRight:                    1.100,
			TrackOrdinal:                         1,
		},
		Race:      race,
		CreatedAt: createdAt,
	}
}
