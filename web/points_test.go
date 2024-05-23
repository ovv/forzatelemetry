package web_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"forzatelemetry/testutils"
	"forzatelemetry/web"
)

type testStreamRacePointsHistoryRun struct {
	code     int
	raceID   string
	errorMsg string
	result   []byte
}

func TestStreamRacePointsHistory(t *testing.T) {
	db := testutils.NewStore("races.yaml", "points.yaml")
	defer db.Close()

	router := web.Router(db, "version")

	runs := map[string]testStreamRacePointsHistoryRun{
		"ok":       {code: 200, raceID: "44e22d85-3883-4552-9ff4-91a7211e0639", result: []byte{0, 0, 0, 0, 7, 0, 0, 0, 13, 0, 0, 240, 66, 24, 1, 7, 0, 0, 0, 13, 0, 0, 112, 67, 24, 2}},
		"missing":  {code: 404, raceID: "44e22d85-3883-4552-9ff4-aaaaaaaaaaaa", errorMsg: "not found"},
		"noPoints": {code: 404, raceID: "df9d1160-4c51-4b94-824f-c6f9cd1dee5e", errorMsg: "not found"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/races/%s/points", run.raceID), nil)
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
				checkSuccess(resp, run, t)
			} else {
				testutils.CheckErrorPayload(resp, run.errorMsg, t)
			}
		})
	}
}

func checkSuccess(resp *httptest.ResponseRecorder, run testStreamRacePointsHistoryRun, t *testing.T) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}
	if !reflect.DeepEqual(data, run.result) {
		t.Errorf("expected %v got %v", run.result, data)
	}
}
