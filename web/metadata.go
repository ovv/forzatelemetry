package web

import (
	"net/http"

	"forzatelemetry/models"
)

func (h Handler) tracksMetadata(w http.ResponseWriter, r *http.Request) {
	tracks := map[int]string{}
	for _, track := range models.Tracks {
		tracks[track.Ordinal] = track.FullName()
	}

	Render(w, r, TracksRenderer{
		Count: len(models.Tracks),
		Items: models.Tracks,
	})
}

type TracksRenderer struct {
	Renderer `json:"-"`

	Count int            `json:"count"`
	Items []models.Track `json:"items"`
}

type TracksTemplateData struct {
	TemplateData `json:"-"`
	Tracks       map[string]models.Track
}

func (rd TracksRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	tracks := make(map[string]models.Track)
	for _, track := range rd.Items {
		tracks[track.FullName()] = track
	}

	return RenderTemplate(r, "tracks.html", TracksTemplateData{
		TemplateData: NewTemplateData(r),
		Tracks:       tracks,
	})
}
