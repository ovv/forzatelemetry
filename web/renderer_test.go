package web_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun/driver/pgdriver"

	"forzatelemetry/testutils"
	"forzatelemetry/web"
)

func TestExplicitErrorRenderer(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		web.Render(w, r, web.NewErrorRenderer(500, "test error", errors.New("test error"), map[string]any{
			"foo": "bar",
		}))
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	resp := testutils.ExecuteRequest(req, router)

	if resp.Code != 500 {
		t.Errorf("expected 500 got %v", resp.Code)
	}

	var response web.ErrorRenderer
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if response.Msg != "test error" {
		t.Errorf("expected test error got %s", response.Msg)
	}

	if !reflect.DeepEqual(response.Details, map[string]any{"foo": "bar"}) {
		t.Errorf("expected foo:bar details got %s", response.Details)
	}
}

type faillingRenderer struct {
}

func (rd *faillingRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	return errors.New("test error")
}

func (rd *faillingRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	return "test error"
}

func TestRenderError(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		web.Render(w, r, &faillingRenderer{})
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	resp := testutils.ExecuteRequest(req, router)

	if resp.Code != 500 {
		t.Errorf("expected 500 got %v", resp.Code)
	}

	var response web.ErrorRenderer
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if response.Msg != "internal error" {
		t.Errorf("expected internal error got %s", response.Msg)
	}

	if !reflect.DeepEqual(response.Details, map[string]any{}) {
		t.Errorf("expected empty details got %s", response.Details)
	}
}

type testStorageErrorRendererRun struct {
	code   int
	err    error
	errMsg string
}

func TestStorageErrorRenderer(t *testing.T) {

	runs := map[string]testStorageErrorRendererRun{
		"notFound": {code: 404, err: sql.ErrNoRows, errMsg: "not found"},
		"internal": {code: 500, err: sql.ErrConnDone, errMsg: "internal error"},
		"pgdrive":  {code: 500, err: pgdriver.Error{}, errMsg: "internal error"},
	}

	for _, run := range runs {
		router := chi.NewRouter()
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			web.Render(w, r, web.StorageErrorRenderer(run.err))
		})

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}

		resp := testutils.ExecuteRequest(req, router)

		if resp.Code != run.code {
			t.Errorf("expected %v got %v", run.code, resp.Code)
		}

		var response web.ErrorRenderer
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("unexpected error %s", err)
		}

		if response.Msg != run.errMsg {
			t.Errorf("expected %v got %v", run.errMsg, response.Msg)
		}

		if !reflect.DeepEqual(response.Details, map[string]any{}) {
			t.Errorf("expected empty details got %s", response.Details)
		}
	}
}
