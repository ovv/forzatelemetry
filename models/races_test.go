package models_test

import (
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/testutils"
)

func TestRace(t *testing.T) {
	point := testutils.Point(testutils.ParseUUID("7f753007-0eda-4aec-8d25-de6ac96220fc"), testutils.ParseTime("2024-09-08T17:39:10Z"), 0)
	session := testutils.ParseUUID("0c5d98f4-3df8-4a60-88d7-a1ecdecea57c")

	race := models.MakeRace(point.TelemetryPoint, session)

	if !race.Paused {
		t.Errorf("expected %v got %v", true, race.Paused)
	}
	if !race.InProgress {
		t.Errorf("expected %v got %v", true, race.InProgress)
	}

	point = testutils.Point(testutils.ParseUUID("7f753007-0eda-4aec-8d25-de6ac96220fc"), testutils.ParseTime("2024-09-08T17:43:10Z"), 0)
	race = race.End(point)

	if !race.Paused {
		t.Errorf("expected %v got %v", true, race.Paused)
	}
	if race.InProgress {
		t.Errorf("expected %v got %v", false, race.InProgress)
	}
}

func TestRaceDetailled(t *testing.T) {
	point := testutils.Point(testutils.ParseUUID("7f753007-0eda-4aec-8d25-de6ac96220fc"), testutils.ParseTime("2024-09-08T17:39:10Z"), 0)
	session := testutils.ParseUUID("0c5d98f4-3df8-4a60-88d7-a1ecdecea57c")

	race := models.MakeRace(point.TelemetryPoint, session)
	raceDetailled := models.MakeRaceDetailled(models.APIRace{Race: race}, nil)

	if raceDetailled.Laps == nil {
		t.Errorf("expected map got %v", nil)
	}

	raceDetailled = models.MakeRaceDetailled(models.APIRace{Race: race}, map[uint16]models.Lap{
		0: {},
	})

	if len(raceDetailled.Laps) != 1 {
		t.Errorf("expected 1 got %v", len(raceDetailled.Laps))
	}
}
