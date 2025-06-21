package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestExecuteRequest(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	rr := ExecuteRequest(req, r)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "pong" {
		t.Errorf("expected body 'pong', got '%s'", rr.Body.String())
	}
}

func TestCheckErrorPayload(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Body.WriteString(`{"error":"test error","details":{}}`)
	CheckErrorPayload(recorder, "test error", t)
}

func TestParseTime(t *testing.T) {
	ts := "2023-01-01T12:00:00.000000000Z"
	parsed := ParseTime(ts)
	expected, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if !parsed.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, parsed)
	}
}

func TestParseUUID(t *testing.T) {
	id := uuid.New().String()
	parsed := ParseUUID(id)
	if parsed.String() != id {
		t.Errorf("expected %s, got %s", id, parsed.String())
	}
}
