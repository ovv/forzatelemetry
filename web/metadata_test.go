package web_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"forzatelemetry/models"
	"forzatelemetry/testutils"
	"forzatelemetry/web"
	"net/http/httptest"
)

func TestTracks(t *testing.T) {
	db := testutils.NewStore()
	defer db.Close()

	router := web.Router(db, "version", "https://localhost")

	req, err := http.NewRequest("GET", "/metadata/tracks", nil)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	resp := testutils.ExecuteRequest(req, router)

	contentType := resp.Header().Get("Content-Type")
	expectedContentType := "application/json"
	if contentType != expectedContentType {
		t.Fatalf("expected %v got %v", expectedContentType, contentType)
	}

	expectedCode := 200
	if resp.Code != expectedCode {
		t.Fatalf("expected %v got %v", expectedCode, resp.Code)
	}

	type data struct {
		Count int            `json:"count"`
		Items []models.Track `json:"items"`
	}

	var d data
	err = json.NewDecoder(resp.Body).Decode(&d)
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if d.Count <= 0 {
		t.Fatalf("negative or zero items count: %v", d.Count)
	}

	expected := "WeatherTech Raceway Laguna Seca - Short Circuit"
	if d.Items[1] != models.Tracks[1] {
		t.Fatalf("expected %v got %v", expected, d.Items[1])
	}
}

func TestTracksRendererHTML(t *testing.T) {
	r := web.TracksRenderer{
		Count: 2,
		Items: models.Tracks[:2],
	}
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	html := r.HTML(w, req)
	if html == "" {
		t.Error("expected non-empty HTML output")
	}
}
