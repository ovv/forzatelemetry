package storage_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"forzatelemetry/models"
	"forzatelemetry/storage"
	"forzatelemetry/testutils"
)

type selectLastPointRun struct {
	race      string
	createdAt time.Time
	errMsg    string
}

func TestSelectLastPoint(t *testing.T) {
	db := testutils.NewStore("points.yaml")
	defer db.Close()

	runs := map[string]selectLastPointRun{
		"ok":      {race: "44e22d85-3883-4552-9ff4-91a7211e0639", createdAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
		"missing": {race: "44e22d85-3883-4552-9ff4-aaaaaaaaaaaa", errMsg: "sql: no rows in result set"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			point, err := db.SelectLastPoint(run.race, context.Background())
			if err != nil {
				if run.errMsg != err.Error() {
					t.Errorf("expected %v got %v", run.errMsg, err.Error())
				}
			} else if run.errMsg != "" {
				t.Errorf("expected %v got %v", run.errMsg, err)
			}

			if run.createdAt != point.CreatedAt.UTC() {
				t.Errorf("expected %v got %v", run.createdAt, point.CreatedAt)
			}
		})
	}
}

type iterPointsRun struct {
	race      string
	createdAt []time.Time
	errMsg    string
}

func TestIterPoints(t *testing.T) {
	db := testutils.NewStore("points.yaml")
	defer db.Close()

	runs := map[string]iterPointsRun{
		"ok": {race: "44e22d85-3883-4552-9ff4-91a7211e0639", createdAt: []time.Time{
			testutils.ParseTime("2024-09-08T17:37:10.000000Z"),
			testutils.ParseTime("2024-09-08T17:39:10.000000Z"),
			testutils.ParseTime("2024-09-08T17:41:10.000000Z"),
		}},
		"missing": {race: "44e22d85-3883-4552-9ff4-aaaaaaaaaaaa", createdAt: []time.Time{{}}, errMsg: "sql: no rows in result set"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			var points []models.Point
			for point, err := range db.IterPoints(run.race, []storage.Where{}, context.Background()) {
				if err != nil {
					if run.errMsg != err.Error() {
						t.Errorf("expected %v got %v", run.errMsg, err.Error())
						return
					}
				} else if run.errMsg != "" {
					t.Errorf("expected %v got %v", run.errMsg, err)
				}
				points = append(points, point)
			}

			if len(run.createdAt) != len(points) {
				t.Errorf("expected %v got %v", len(run.createdAt), len(points))
			}

			for i, point := range points {
				if point.CreatedAt.UTC() != run.createdAt[i] {
					t.Errorf("expected %v got %v", run.createdAt[i], point.CreatedAt.UTC())
				}
			}
		})
	}
}

func TestIterPointsStopEarly(t *testing.T) {
	db := testutils.NewStore("points.yaml")
	defer db.Close()

	race := "44e22d85-3883-4552-9ff4-91a7211e0639"
	count := 0

	var point models.Point
	for p, err := range db.IterPoints(race, []storage.Where{}, context.Background()) {
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		if count == 0 {
			count = count + 1
			continue
		}
		point = p
		break
	}

	expected := testutils.ParseTime("2024-09-08T17:39:10.000000Z")
	if point.CreatedAt.UTC() != expected {
		t.Errorf("expected %v got %v", expected, point.CreatedAt.UTC())
	}
}

type selectLapsRun struct {
	race   string
	laps   map[uint16]models.Lap
	errMsg string
}

func TestSelectLaps(t *testing.T) {
	db := testutils.NewStore("points.yaml")
	defer db.Close()

	runs := map[string]selectLapsRun{
		"ok": {race: "982e6f1d-efe2-4b67-b420-67c08e705994", laps: map[uint16]models.Lap{
			0: {LapNumber: 0, LapTime: 60, FinishedAt: testutils.ParseTime("2024-09-08T17:38:10.000000Z"), RacePosition: 9, RaceTime: 60},
			1: {LapNumber: 1, LapTime: 55, FinishedAt: testutils.ParseTime("2024-09-08T17:40:10.000000Z"), RacePosition: 7, RaceTime: 180},
			2: {LapNumber: 2, LapTime: 64, FinishedAt: testutils.ParseTime("2024-09-08T17:42:10.000000Z"), RacePosition: 5, RaceTime: 300},
		}},
		"missing": {race: "44e22d85-3883-4552-9ff4-aaaaaaaaaaaa", laps: map[uint16]models.Lap{}, errMsg: "sql: no rows in result set"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			laps, err := db.SelectLaps(run.race, context.Background())
			if err != nil {
				if run.errMsg != err.Error() {
					t.Errorf("expected %v got %v", run.errMsg, err.Error())
				}
			} else if run.errMsg != "" {
				t.Errorf("expected %v got %v", run.errMsg, err)
			}

			for i, lap := range laps {
				lap.FinishedAt = lap.FinishedAt.UTC()
				laps[i] = lap
			}

			if !reflect.DeepEqual(run.laps, laps) {
				t.Errorf("expected %+v got %+v", run.laps, laps)
			}
		})
	}
}

func TestInsertPoints(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	points := []models.Point{
		testutils.Point(testutils.ParseUUID("a6996827-6699-4206-8168-4584cb2176e2"), time.Now(), 1),
		testutils.Point(testutils.ParseUUID("a6996827-6699-4206-8168-4584cb2176e2"), time.Now(), 1),
		testutils.Point(testutils.ParseUUID("06874f47-0f56-44bd-9dc2-b12e5cbfed5e"), time.Now(), 1),
	}

	err := db.InsertPoints(points, context.Background())
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	var result []models.Point
	for point, err := range db.IterPoints("a6996827-6699-4206-8168-4584cb2176e2", []storage.Where{}, context.Background()) {
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		result = append(result, point)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 got %v", len(result))
	}
}
