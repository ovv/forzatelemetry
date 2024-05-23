package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Race struct {
	bun.BaseModel

	ID        uuid.UUID `bun:"type:uuid,unique,pk"`
	SessionID uuid.UUID `bun:"type:uuid" json:"sessionID"`

	Paused     bool `json:"paused"`
	InProgress bool `json:"inProgress"`

	StartedAt  time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt" bun:",nullzero"`

	Car                 int32 `json:"car"`
	CarClass            int32 `json:"carClass"`
	CarPerformanceIndex int32 `json:"carPerformanceIndex"`

	Track int32 `json:"track"`

	BestLap          float32 `json:"bestLap"`
	RaceTime         float32 `json:"raceTime"`
	Position         uint8   `json:"position"`
	DistanceTraveled float32 `json:"distanceTraveled"`
}

func (r Race) Update(point Point) Race {
	r.BestLap = point.BestLap
	r.RaceTime = point.CurrentRaceTime
	r.Position = point.RacePosition
	r.DistanceTraveled = point.DistanceTraveled
	return r
}

func (r Race) End(point Point) Race {
	r.InProgress = false
	r.FinishedAt = point.CreatedAt
	return r.Update(point)
}

func MakeRace(p TelemetryPoint, sessionId uuid.UUID) Race {
	race := Race{
		ID:                  uuid.New(),
		Paused:              p.OnTrack == 0,
		InProgress:          true,
		SessionID:           sessionId,
		Car:                 p.CarOrdinal,
		CarClass:            p.CarClass,
		CarPerformanceIndex: p.CarPerformanceIndex,
		Track:               p.TrackOrdinal,
	}
	return race
}

type APIRace struct {
	bun.BaseModel `bun:"table:races,alias:races"`
	Race

	TrackMetadata    Track    `json:"trackMetadata" bun:"rel:belongs-to,join:track=ordinal"`
	CarMetadata      Car      `json:"carMetadata" bun:"rel:belongs-to,join:car=ordinal"`
	CarClassMetadata CarClass `json:"carClassMetadata" bun:"rel:belongs-to,join:car_class=id"`
	Dashboard        string   `json:"dashboard"`
}

type APIRaceDetailled struct {
	APIRace
	Laps map[uint16]Lap `json:"laps"`
}

func MakeRaceDetailled(race APIRace, laps map[uint16]Lap) APIRaceDetailled {
	if laps == nil {
		laps = map[uint16]Lap{}
	}
	return APIRaceDetailled{
		APIRace: race,
		Laps:    laps,
	}
}

type Lap struct {
	LapNumber    uint16    `bun:"lap_number" json:"lapNumber"`
	LapTime      float32   `bun:"current_lap" json:"lapTime"`
	FinishedAt   time.Time `bun:"created_at" json:"finishedAt"`
	RacePosition uint8     `bun:"race_position" json:"racePosition"`
	RaceTime     float32   `bun:"current_race_time" json:"raceTime"`
}
