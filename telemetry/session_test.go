package telemetry_test

import (
	"context"
	"testing"
	"time"

	"forzatelemetry/models"
	"forzatelemetry/storage"
	"forzatelemetry/telemetry"
	"forzatelemetry/testutils"
)

func assertNoRace(t *testing.T, db *storage.Store) {
	races, count, err := db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 got %v", count)
	}

	if len(races) != 0 {
		t.Errorf("expected 0 got %v", len(races))
	}
}

func TestSessionCreateRace(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:    1,
		CurrentLap: 1,
	})

	_, count, err := db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 got %v", count)
	}
}

func TestSessionPauseRace(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:    1,
		CurrentLap: 1,
	})

	races, _, err := db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if races[0].Paused {
		t.Errorf("expected false got %v", races[0].Paused)
	}

	session.Add(models.TelemetryPoint{
		OnTrack:    0,
		CurrentLap: 1,
	})

	races, _, err = db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if !races[0].Paused {
		t.Errorf("expected true got %v", races[0].Paused)
	}

	session.Add(models.TelemetryPoint{
		OnTrack:    1,
		CurrentLap: 1,
	})

	races, _, err = db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if races[0].Paused {
		t.Errorf("expected false got %v", races[0].Paused)
	}
}

func TestSessionPauseNoRace(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:    0,
		CurrentLap: 1,
	})
	assertNoRace(t, db)
}

func TestSessionNoCurrentLap(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:    1,
		CurrentLap: 0,
	})
	assertNoRace(t, db)
}

func TestSessionCloseNoRace(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:    0,
		CurrentLap: 0,
	})

	session.Close()
	assertNoRace(t, db)
}

func TestSessionCloseRaceUpdate(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 1000,
	})

	races, _, err := db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if races[0].RaceTime != 0 {
		t.Errorf("expected 0 got %v", races[0].RaceTime)
	}

	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 2000,
	})

	// sleep a bit to order points
	time.Sleep(10 * time.Microsecond)

	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 3000,
	})
	session.Close()

	races, _, err = db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if races[0].RaceTime != 3000 {
		t.Errorf("expected 3000 got %v", races[0].RaceTime)
	}
}

func TestSessionNewRace(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 1000,
	})

	races, count, err := db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 got %v", count)
	}

	initialRace := races[0]

	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 2000,
	})

	// sleep a bit to order points
	time.Sleep(10 * time.Microsecond)

	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 3000,
	})

	// sleep a bit to order points
	time.Sleep(10 * time.Microsecond)

	session.Add(models.TelemetryPoint{
		OnTrack:         1,
		CurrentLap:      1,
		CurrentRaceTime: 0.5,
	})

	races, count, err = db.SelectRaces(nil, 0, context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 got %v", count)
	}

	if initialRace.ID != races[0].ID {
		t.Errorf("expected %v got %v", initialRace.ID, races[0].ID)
	}
	if initialRace.RaceTime != 0 {
		t.Errorf("expected 0 got %v", initialRace.RaceTime)
	}
	if races[0].RaceTime != 3000 {
		t.Errorf("expected 3000 got %v", races[0].RaceTime)
	}
}

func TestSessionChecpoi(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	session := telemetry.NewSession(db)
	err := session.Checkpoint()
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
}
