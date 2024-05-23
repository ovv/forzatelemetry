package web_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/testutils"
	"forzatelemetry/web"
)

type testGetRacesRun struct {
	code       int
	countTotal int
	countItems int
	params     string
	errorMsg   string
}

type getRacesResponse struct {
	Count int              `json:"count"`
	Items []models.APIRace `json:"items"`
}

func TestGetRaces(t *testing.T) {
	db := testutils.NewStore("races.yaml")
	defer db.Close()

	router := web.Router(db, "version")

	runs := map[string]testGetRacesRun{
		"empty":         {code: 200, countTotal: 10, countItems: 10, params: ""},
		"inProgress":    {code: 200, countTotal: 1, countItems: 1, params: "filter=inProgress:eq:true"},
		"notInProgress": {code: 200, countTotal: 9, countItems: 9, params: "filter=inProgress:eq:false"},
		"wrongFilter":   {code: 400, countTotal: 0, countItems: 0, params: "filter=inProgress:eq:aaa", errorMsg: "invalid filter 'inProgress:eq:aaa': invalid value: invalid syntax"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/races?%s", run.params), nil)
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}

			resp := testutils.ExecuteRequest(req, router)

			if resp.Code != run.code {
				t.Log(resp.Body)
				t.Errorf("expected %v got %v", run.code, resp.Code)
				return
			}

			if resp.Code == 200 {
				testGetRacesCheckSuccess(resp, run, t)
			} else {
				testutils.CheckErrorPayload(resp, run.errorMsg, t)
			}
		})
	}
}

func testGetRacesCheckSuccess(resp *httptest.ResponseRecorder, run testGetRacesRun, t *testing.T) {
	var respData getRacesResponse
	err := json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if respData.Count != run.countTotal {
		t.Errorf("expected %v got %v", run.countTotal, respData.Count)
	}

	if len(respData.Items) != run.countItems {
		t.Errorf("expected %v got %v", run.countItems, len(respData.Items))
	}
}

type testGetRaceRun struct {
	code     int
	raceID   string
	race     models.APIRaceDetailled
	errorMsg string
}

func makeRaceDetailled(id string, sessionID string, startedAt string, finishedAt string, laps map[uint16]models.Lap) models.APIRaceDetailled {
	if laps == nil {
		laps = make(map[uint16]models.Lap)
	}

	return models.APIRaceDetailled{
		APIRace: models.APIRace{
			Race: models.Race{
				ID:         testutils.ParseUUID(id),
				SessionID:  testutils.ParseUUID(sessionID),
				StartedAt:  testutils.ParseTime(startedAt),
				FinishedAt: testutils.ParseTime(finishedAt),
				Car:        100,
			},
		},
		Laps: laps,
	}
}

func TestGetRace(t *testing.T) {
	db := testutils.NewStore("races.yaml", "points.yaml")
	defer db.Close()

	router := web.Router(db, "version")

	runs := map[string]testGetRaceRun{
		"exist": {
			code:   200,
			raceID: "44e22d85-3883-4552-9ff4-91a7211e0639",
			race: makeRaceDetailled("44e22d85-3883-4552-9ff4-91a7211e0639", "665078b0-1130-48a9-8a35-0e7cbfd7704c", "2024-09-07T17:37:10Z", "2024-09-08T17:37:10Z", map[uint16]models.Lap{
				0: {LapNumber: 0, RaceTime: 0, FinishedAt: testutils.ParseTime("2024-09-08T17:37:10Z")},
				1: {LapNumber: 1, RaceTime: 120, FinishedAt: testutils.ParseTime("2024-09-08T17:39:10Z")},
				2: {LapNumber: 2, RaceTime: 240, FinishedAt: testutils.ParseTime("2024-09-08T17:41:10Z")},
			}),
		},
		"missing":   {code: 404, raceID: "44e22d85-3883-4552-9ff4-aaaaaaaaaaaa", errorMsg: "not found"},
		"invalidID": {code: 404, raceID: "aaaa", errorMsg: "not found"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/races/%s", run.raceID), nil)
			if err != nil {
				t.Errorf("unexpected error %s", err)
			}

			resp := testutils.ExecuteRequest(req, router)

			if resp.Code != run.code {
				t.Log(resp.Body)
				t.Errorf("expected %v got %v", run.code, resp.Code)
				return
			}

			if resp.Code == 200 {
				testGetRaceCheckSuccess(resp, run, t)
			} else {
				testutils.CheckErrorPayload(resp, run.errorMsg, t)
			}
		})
	}
}

type getRaceResponse struct {
	Race models.APIRaceDetailled `json:"race"`
}

func testGetRaceCheckSuccess(resp *httptest.ResponseRecorder, run testGetRaceRun, t *testing.T) {
	var respData getRaceResponse
	err := json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if !reflect.DeepEqual(respData.Race, run.race) {
		t.Errorf("expected %+v got %+v", run.race, respData.Race)
	}
}
