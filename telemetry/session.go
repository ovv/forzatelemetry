package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"forzatelemetry/models"
	"forzatelemetry/storage"
)

// Allocate the array once. We receive 60 points per second maximum, for 5s we need a capacity of 300. We take a small margin to make sure we don't need to grow the array
const POINT_ARRAY_LENGTH = 350

const SESSION_CHANNEL_SIZE = 60

var EMPTY_UUID = uuid.UUID{}

type Session struct {
	ID uuid.UUID
	db *storage.Store

	points []models.Point
	race   models.Race
}

func NewSession(db *storage.Store) *Session {
	session := &Session{
		ID:     uuid.New(),
		db:     db,
		points: make([]models.Point, 0, POINT_ARRAY_LENGTH),
	}
	slog.Info("new session", "session", session.ID)
	return session
}

func (s *Session) Close() error {
	slog.Info("closing session", "session", s.ID)
	if s.race.ID == EMPTY_UUID {
		return nil
	}
	err := s.Checkpoint()
	if err != nil {
		slog.Error("failed checkpointing before closing session", "error", err, "session", s.ID)
	}
	err = s.endRace()
	return err
}

func (s *Session) Add(point models.TelemetryPoint) error {
	if point.OnTrack == 0 {
		return s.pause()
	}

	if point.CurrentLap == 0 {
		return nil
	}

	var err error
	if s.race.ID == EMPTY_UUID {
		s.race = models.MakeRace(point, s.ID)
		slog.Info("new race", "id", s.race.ID, "session", s.ID)
		err = s.saveRace()
	} else if s.isNewRace(point) {
		slog.Info("finished race", "race", s.race.ID, "session", s.ID)
		err = s.Checkpoint()
		if err != nil {
			return err
		}
		err = s.endRace()
		if err != nil {
			return err
		}

		newRace := models.MakeRace(point, s.ID)
		slog.Info("new race", "id", newRace.ID, "session", s.ID)
		err = s.db.UpsertRaces(context.Background(), s.race, newRace)
		if err != nil {
			return err
		}
		s.race = newRace
	} else if s.race.Paused {
		err = s.unpause()
	}

	s.points = append(s.points, models.Point{TelemetryPoint: point, Race: s.race.ID, CreatedAt: time.Now()})
	s.race.RaceTime = point.CurrentRaceTime
	return err
}

func (s *Session) pause() error {
	if (s.race.ID != EMPTY_UUID) && !s.race.Paused {
		slog.Info("pausing", "race", s.race.ID)
		err := s.Checkpoint()
		if err != nil {
			return fmt.Errorf("failed to pause race: %w", err)
		}
		s.race.Paused = true
		err = s.saveRace()
		if err != nil {
			return fmt.Errorf("failed to pause race: %w", err)
		}
	}
	return nil
}

func (s *Session) unpause() error {
	slog.Info("unpausing", "race", s.race.ID)
	s.race.Paused = false
	err := s.saveRace()
	if err != nil {
		return fmt.Errorf("failed to unpause race: %w", err)
	}
	return nil
}

func (s *Session) Checkpoint() error {
	// No new points, nothing to save
	if len(s.points) == 0 {
		return nil
	}
	start := time.Now()

	s.race = s.race.Update(s.points[len(s.points)-1])
	s.saveRace()

	err := s.db.InsertPoints(s.points, context.Background())
	if err != nil {
		return fmt.Errorf("failed checkpointing race %s: %w", s.race.ID, err)
	}
	s.points = s.points[:0]
	slog.Debug("checkpoint", "race", s.race.ID, "points", len(s.points), "cap", cap(s.points), "duration", time.Since(start))
	return nil
}

func (s *Session) isNewRace(point models.TelemetryPoint) bool {
	// Figuring out when a "race" ends or start is tricky as it heavily depends on what you consider a race and the type of lobby (solo vs multiplayer).
	// We can mostly based ourselve on the currentRaceTime. It starts a 0 and if it ever goes backward we can expect it to be a new race.
	// But that's not enough, there are some exceptions were the time goes backward :'(
	// * rewind in solo play
	// * re-joining a multiplayer session
	// In this case we hope that at least one lap was done, then the lastLap value is set.
	return point.CurrentRaceTime < 1 && point.CurrentRaceTime < s.race.RaceTime && point.LastLap == 0
}

func (s *Session) saveRace() error {
	err := s.db.UpsertRaces(context.Background(), s.race)
	if err != nil {
		return fmt.Errorf("failed to upsert race: %w", err)
	}
	return err
}

func (s *Session) endRace() error {
	var err error
	point, err := s.db.SelectLastPoint(s.race.ID.String(), context.Background())
	if err != nil {
		return fmt.Errorf("failed to read last point of race: %w", err)
	}

	s.race = s.race.End(point)
	return s.saveRace()
}
