package storage_test

import (
	"context"
	"reflect"
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/storage"
	"forzatelemetry/testutils"

	"github.com/google/uuid"
)

type selectRacesRun struct {
	count  int
	where  []storage.Where
	offset int
	result []uuid.UUID
}

func TestSelectRaces(t *testing.T) {
	db := testutils.NewStore("races.yaml")
	defer db.Close()

	runs := map[string]selectRacesRun{
		"all": {count: 10, where: []storage.Where{}, offset: 0, result: []uuid.UUID{
			testutils.ParseUUID("db1e4ffd-9476-48fc-b611-154cfc9e7c02"),
			testutils.ParseUUID("b1856e96-66b4-410b-a4b4-8428df3af2cb"),
			testutils.ParseUUID("3cb31256-f9fb-481b-90bc-9b7440441105"),
			testutils.ParseUUID("9d311ab7-9236-42e6-9287-62f9b41fe1d1"),
			testutils.ParseUUID("3b3e9041-dac2-4411-8785-1c3546098ef6"),
			testutils.ParseUUID("ad28a390-ac89-4969-89a2-a0f7e03b4b2b"),
			testutils.ParseUUID("54221549-d8cc-4726-862b-fb8cf92b4677"),
			testutils.ParseUUID("5d8f6c20-e31e-4b41-ab97-aa07f3bef3fd"),
			testutils.ParseUUID("df9d1160-4c51-4b94-824f-c6f9cd1dee5e"),
			testutils.ParseUUID("44e22d85-3883-4552-9ff4-91a7211e0639"),
		}},
		"offset": {count: 10, where: []storage.Where{}, offset: 5, result: []uuid.UUID{
			testutils.ParseUUID("ad28a390-ac89-4969-89a2-a0f7e03b4b2b"),
			testutils.ParseUUID("54221549-d8cc-4726-862b-fb8cf92b4677"),
			testutils.ParseUUID("5d8f6c20-e31e-4b41-ab97-aa07f3bef3fd"),
			testutils.ParseUUID("df9d1160-4c51-4b94-824f-c6f9cd1dee5e"),
			testutils.ParseUUID("44e22d85-3883-4552-9ff4-91a7211e0639"),
		}},
		"whereIn": {count: 3, where: []storage.Where{
			{Column: "car", Operator: "IN", Value: []int{102, 104, 106, 200}},
		}, offset: 0, result: []uuid.UUID{
			testutils.ParseUUID("9d311ab7-9236-42e6-9287-62f9b41fe1d1"),
			testutils.ParseUUID("ad28a390-ac89-4969-89a2-a0f7e03b4b2b"),
			testutils.ParseUUID("5d8f6c20-e31e-4b41-ab97-aa07f3bef3fd"),
		}},
		"whereEq": {count: 2, where: []storage.Where{
			{Column: "car", Operator: ">", Value: 107},
		}, offset: 0, result: []uuid.UUID{
			testutils.ParseUUID("db1e4ffd-9476-48fc-b611-154cfc9e7c02"),
			testutils.ParseUUID("b1856e96-66b4-410b-a4b4-8428df3af2cb"),
		}},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			races, count, err := db.SelectRaces(run.where, run.offset, context.Background())
			if err != nil {
				t.Fatal("unexpected error")
			}
			if count != run.count {
				t.Errorf("expected %v got %v", run.count, count)
			}
			var racesID []uuid.UUID
			for _, race := range races {
				racesID = append(racesID, race.ID)
			}
			if !reflect.DeepEqual(racesID, run.result) {
				t.Errorf("expected %v got %v", run.result, racesID)
			}
		})
	}
}

type selectRaceRun struct {
	id     string
	result string
	errMsg string
}

func TestSelectRace(t *testing.T) {
	db := testutils.NewStore("races.yaml")
	defer db.Close()

	runs := map[string]selectRaceRun{
		"ok":      {id: "df9d1160-4c51-4b94-824f-c6f9cd1dee5e", result: "df9d1160-4c51-4b94-824f-c6f9cd1dee5e"},
		"missing": {id: "df9d1160-4c51-4b94-824f-aaaaaaaaaaaa", result: "00000000-0000-0000-0000-000000000000", errMsg: "sql: no rows in result set"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			race, err := db.SelectRace(run.id, context.Background())
			if err != nil {
				if run.errMsg != err.Error() {
					t.Errorf("expected %v got %v", run.errMsg, err.Error())
				}
			} else if run.errMsg != "" {
				t.Errorf("expected %v got %v", run.errMsg, err)
			}

			if race.ID != testutils.ParseUUID(run.result) {
				t.Errorf("expected %v got %v", run.result, race.ID)
			}
		})
	}
}

type selectRaceLapsRun struct {
	id     string
	result models.APIRaceDetailled
	errMsg string
}

func TestSelectRaceLaps(t *testing.T) {
	db := testutils.NewStore("races.yaml", "points.yaml")
	defer db.Close()

	runs := map[string]selectRaceLapsRun{
		"ok": {id: "982e6f1d-efe2-4b67-b420-67c08e705994", result: models.APIRaceDetailled{
			APIRace: models.APIRace{
				Race: models.Race{
					ID:        testutils.ParseUUID("982e6f1d-efe2-4b67-b420-67c08e705994"),
					SessionID: testutils.ParseUUID("b1f08eb3-544c-46f9-a005-aaff6c41c215"),
					Car:       108,
					StartedAt: testutils.ParseTime("2024-09-16T17:37:10.0Z"),
				},
			},
			Laps: map[uint16]models.Lap{
				0: {LapNumber: 0, LapTime: 60, FinishedAt: testutils.ParseTime("2024-09-08T17:38:10.000000Z"), RacePosition: 9, RaceTime: 60},
				1: {LapNumber: 1, LapTime: 55, FinishedAt: testutils.ParseTime("2024-09-08T17:40:10.000000Z"), RacePosition: 7, RaceTime: 180},
				2: {LapNumber: 2, LapTime: 64, FinishedAt: testutils.ParseTime("2024-09-08T17:42:10.000000Z"), RacePosition: 5, RaceTime: 300},
			},
		}},
		"missing": {id: "df9d1160-4c51-4b94-824f-aaaaaaaaaaaa", result: models.APIRaceDetailled{}, errMsg: "sql: no rows in result set"},
		"noLaps": {id: "3cb31256-f9fb-481b-90bc-9b7440441105", result: models.APIRaceDetailled{
			APIRace: models.APIRace{
				Race: models.Race{
					ID:         testutils.ParseUUID("3cb31256-f9fb-481b-90bc-9b7440441105"),
					SessionID:  testutils.ParseUUID("869696cb-d3da-4ed9-a353-111f75cccf77"),
					Car:        107,
					StartedAt:  testutils.ParseTime("2024-09-14T17:37:10.0Z"),
					FinishedAt: testutils.ParseTime("2024-09-15T17:37:10.0Z"),
				},
			},
			Laps: make(map[uint16]models.Lap),
		},
		},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			race, err := db.SelectRaceLaps(run.id, context.Background())
			if err != nil {
				if run.errMsg != err.Error() {
					t.Errorf("expected %v got %v", run.errMsg, err.Error())
				}
			} else if run.errMsg != "" {
				t.Errorf("expected %v got %v", run.errMsg, err)
			}

			for i, lap := range race.Laps {
				lap.FinishedAt = lap.FinishedAt.UTC()
				race.Laps[i] = lap
			}

			race.StartedAt = race.StartedAt.UTC()
			race.FinishedAt = race.FinishedAt.UTC()
			if !reflect.DeepEqual(race, run.result) {
				t.Errorf("expected %+v got %+v", run.result, race)
			}
		})
	}
}

type upsertRacesRun struct {
	races  []models.Race
	update []models.Race
	result []models.Race
	errMsg string
}

func TestUpsertRaces(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	runs := map[string]upsertRacesRun{
		"noRaces": {errMsg: "bun: Insert(empty *reflect.rtype)"},
		"insert": {
			result: []models.Race{{ID: testutils.ParseUUID("6fc02281-be30-4e6b-b4bc-4a36ae5ec933"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
			races:  []models.Race{{ID: testutils.ParseUUID("6fc02281-be30-4e6b-b4bc-4a36ae5ec933"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
		},
		"update": {
			result: []models.Race{{ID: testutils.ParseUUID("894b1a37-6510-42ed-a2fa-47a8bde0939e"), Position: 3, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
			races:  []models.Race{{ID: testutils.ParseUUID("894b1a37-6510-42ed-a2fa-47a8bde0939e"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
			update: []models.Race{{ID: testutils.ParseUUID("894b1a37-6510-42ed-a2fa-47a8bde0939e"), Position: 3, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
		},
		"updateInvalidField": {
			result: []models.Race{{ID: testutils.ParseUUID("b364576c-dedc-497f-a281-6fc942f091bf"), Car: 200, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
			races:  []models.Race{{ID: testutils.ParseUUID("b364576c-dedc-497f-a281-6fc942f091bf"), Car: 200, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
			update: []models.Race{{ID: testutils.ParseUUID("b364576c-dedc-497f-a281-6fc942f091bf"), Car: 300, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")}},
		},
		"insertMultiple": {
			result: []models.Race{
				{ID: testutils.ParseUUID("c2074981-5e3a-490a-883d-487ca097ddd6"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
				{ID: testutils.ParseUUID("93670cd5-f0bd-4c19-a75c-5ca0883155ba"), Position: 3, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
			},
			races: []models.Race{
				{ID: testutils.ParseUUID("c2074981-5e3a-490a-883d-487ca097ddd6"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
				{ID: testutils.ParseUUID("93670cd5-f0bd-4c19-a75c-5ca0883155ba"), Position: 3, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
			},
		},
		"updateMultiple": {
			result: []models.Race{
				{ID: testutils.ParseUUID("91589f7b-fb17-4c85-9099-28006c27fd4e"), Position: 20, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
				{ID: testutils.ParseUUID("00d95823-40e3-446f-889e-7500e5b2494a"), Position: 30, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
			},
			races: []models.Race{
				{ID: testutils.ParseUUID("91589f7b-fb17-4c85-9099-28006c27fd4e"), Position: 2, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
				{ID: testutils.ParseUUID("00d95823-40e3-446f-889e-7500e5b2494a"), Position: 3, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
			},
			update: []models.Race{
				{ID: testutils.ParseUUID("91589f7b-fb17-4c85-9099-28006c27fd4e"), Position: 20, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
				{ID: testutils.ParseUUID("00d95823-40e3-446f-889e-7500e5b2494a"), Position: 30, StartedAt: testutils.ParseTime("2024-09-08T17:41:10.000000Z")},
			},
		},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			err := db.UpsertRaces(context.Background(), run.races...)
			if err != nil {
				if run.errMsg != err.Error() {
					t.Errorf("expected %v got %v", run.errMsg, err.Error())
				}
			} else if run.errMsg != "" {
				t.Errorf("expected %v got %v", run.errMsg, err)
			}

			if len(run.result) == 0 {
				return
			}

			if len(run.update) != 0 {
				err = db.UpsertRaces(context.Background(), run.update...)
				if err != nil {
					t.Errorf("unexpected error %s", err)
				}
			}

			for _, result := range run.result {
				race, err := db.SelectRace(result.ID.String(), context.Background())
				if err != nil {
					t.Errorf("unexpected error %s", err)
				}

				race.Race.StartedAt = race.StartedAt.UTC()
				if !reflect.DeepEqual(race.Race, result) {
					t.Errorf("expected %+v got %+v", result, race.Race)
				}
			}
		})
	}
}
