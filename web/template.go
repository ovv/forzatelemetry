package web

import (
	"fmt"
	"forzatelemetry/models"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

var TMPL = buildTemplate()

func buildTemplate() map[string]*template.Template {
	layout := layoutTmpl()

	entries, err := TemplateFS.ReadDir("html")
	if err != nil {
		panic(err)
	}

	tmpl := make(map[string]*template.Template)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		tmpl[entry.Name()] = template.Must(
			template.Must(layout.Clone()).ParseFS(TemplateFS, fmt.Sprintf("html/%s", entry.Name())),
		)
	}
	return tmpl
}

func layoutTmpl() *template.Template {
	root := template.New("root")
	root = root.Funcs(template.FuncMap{
		"formatAsDuration":   formatAsDuration,
		"formatAsDistance":   formatAsDistance,
		"formatTimeAsFilter": formatTimeAsFilter,
		"NewRaceCardTitle":   NewRaceCardTitle,
	},
	)
	return template.Must(root.ParseFS(TemplateFS, "html/layout/*.html"))
}

func formatAsDuration(data float32) string {
	return time.Duration(time.Duration(int64(data*1000)) * time.Millisecond).String()
}

func formatAsDistance(data float32) string {
	return fmt.Sprintf("%.3fkm", data/1000)
}

func formatTimeAsFilter(data time.Time) int64 {
	// Add 1 to make sure it's bigger than the current race
	return data.UnixMilli() + 1
}

type RaceCardTitle struct {
	Race models.APIRace
	TZ   *time.Location
}

func NewRaceCardTitle(race models.APIRace, TZ *time.Location) RaceCardTitle {
	return RaceCardTitle{
		Race: race,
		TZ:   TZ,
	}
}

func (rc RaceCardTitle) StartedAt() string {
	return rc.Race.StartedAt.In(rc.TZ).Format(time.Stamp)
}

func (rc RaceCardTitle) FinishedAt() string {
	return rc.Race.FinishedAt.In(rc.TZ).Format(time.Stamp)
}

type TemplateData struct {
	HTMX bool
	TZ   *time.Location
}

func NewTemplateData(r *http.Request) TemplateData {
	tz, err := time.LoadLocation(r.Header.Get("HX-Timezone"))
	if err != nil {
		slog.Warn("failed loading location", "tz", tz)
		tz = time.UTC
	}

	return TemplateData{
		HTMX: r.Header.Get("HX-Request") == "true",
		TZ:   tz,
	}
}
