package web

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"forzatelemetry/models"
)

var RacesFilters = []Filter{
	MakeFilter("inProgress", "races.in_progress", "bool", []string{"eq", "neq"}, "inProgress:eq:true"),
	MakeFilter("paused", "races.paused", "bool", []string{"eq", "neq"}, "paused:neq:true"),
	MakeFilter("carClass", "races.car_class", "[]int", []string{"in"}, "carClass:in:1,2"),
	MakeFilter("carPI", "races.car_performance_index", "int32", []string{"eq", "neq", "gt", "ge", "lt", "le"}, "carPI:gt:100"),
	MakeFilter("track", "races.track", "int32", []string{"eq", "neq"}, "track:eq:2"),
	MakeFilter("startedAt", "races.started_at", "time", []string{"gt", "lt"}, "startedAt:gt:1725479276147"),
	MakeFilter("finishedAt", "races.finished_at", "time", []string{"gt", "lt"}, "finishedAt:gt:1725479276147"),
}

func (h *Handler) races(w http.ResponseWriter, r *http.Request) {
	var err error
	httpParam := r.URL.Query()

	filters, errRd := ParseFilters(httpParam, RacesFilters, nil)
	if errRd != nil {
		Render(w, r, errRd)
		return
	}

	races, count, err := h.db.SelectRaces(filters, 0, r.Context(), h.dashboardBaseUrl)
	if err != nil {
		Render(w, r, StorageErrorRenderer(err))
		return
	}

	Render(w, r, RacesRenderer{Count: count, Items: races})
}

type RacesRenderer struct {
	Renderer `json:"-"`

	Count int              `json:"count"`
	Items []models.APIRace `json:"items"`
}

func (rd RacesRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	td := RacesTemplateData{
		TemplateData: NewTemplateData(r),
		Count:        rd.Count,
		Items:        rd.Items,
		Polling:      r.Header.Get("HX-Polling"),
	}

	// if we are polling for new races and there aren't any, keep the old HTML.
	// easy way to get keep the same `started_at` filter.
	if len(rd.Items) == 0 && td.Polling == "new" {
		w.Header().Set("HX-Reswap", "none")
	}

	return RenderTemplate(r, "races.html", td)
}

type RacesTemplateData struct {
	TemplateData

	Count   int
	Items   []models.APIRace
	Polling string
}

func (td RacesTemplateData) PollNew() bool {
	return (td.Polling == "new" || td.Polling == "all")
}

func (td RacesTemplateData) PollNewTime() int64 {
	if len(td.Items) > 0 {
		return td.Items[0].StartedAt.UnixMilli() + 1
	} else {
		return time.Now().UnixMilli()
	}
}

func (td RacesTemplateData) PollPrevious(index int) bool {
	if (td.Polling == "previous" || td.Polling == "all") && index+1 == len(td.Items) {
		return true
	}
	return false
}

func (td RacesTemplateData) PollPreviousTime() int64 {
	return td.Items[len(td.Items)-1].StartedAt.UnixMilli()
}

func (h *Handler) race(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	raceDetail, err := h.db.SelectRaceLaps(id, r.Context(), h.dashboardBaseUrl)
	if err != nil {
		Render(w, r, StorageErrorRenderer(err))
		return
	}

	Render(w, r, RaceResponse{Race: raceDetail, TemplateData: NewTemplateData(r)})
}

type RaceResponse struct {
	TemplateData `json:"-"`
	Renderer     `json:"-"`

	Race models.APIRaceDetailled `json:"race"`
}

func (rd RaceResponse) HTML(w http.ResponseWriter, r *http.Request) string {
	return RenderTemplate(r, "race.html", rd)
}
