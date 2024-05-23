package web_test

import (
	net "net/http"
	"reflect"
	"testing"

	"forzatelemetry/testutils"
	http "forzatelemetry/web"
)

type testAPIRun struct {
	code     int
	method   string
	path     string
	errorMsg string
}

func TestApi(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	router := http.Router(db, "version", "https://localhost")

	runs := map[string]testAPIRun{
		"index":       {code: 200, method: "GET", path: "/"},
		"races":       {code: 200, method: "GET", path: "/races"},
		"wrongMethod": {code: 405, method: "POST", path: "/races", errorMsg: "not allowed"},
		"missing":     {code: 404, method: "GET", path: "/missing", errorMsg: "not found"},
	}

	for name, run := range runs {
		t.Run(name, func(t *testing.T) {
			req, err := net.NewRequest(run.method, run.path, nil)
			if err != nil {
				t.Fatalf("unexpected error %s", err)
			}

			resp := testutils.ExecuteRequest(req, router)

			if resp.Code != run.code {
				t.Errorf("expected %v got %v", run.code, resp.Code)
				return
			}

			accessControlHeader := resp.Header()["Access-Control-Allow-Origin"]
			expectedAccessControlHeader := []string{"*"}
			if !reflect.DeepEqual(accessControlHeader, expectedAccessControlHeader) {
				t.Errorf("expected %v got %v", expectedAccessControlHeader, accessControlHeader)
			}

			if resp.Code == 200 || resp.Code == 307 {
				//
			} else {
				testutils.CheckErrorPayload(resp, run.errorMsg, t)
			}
		})
	}
}
