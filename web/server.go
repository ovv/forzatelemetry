package web

import (
	"net/http"

	"forzatelemetry/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router(db *storage.Store, revision string) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)

	hdlr := &Handler{db: db, revision: revision}

	router.Get("/ping", hdlr.pong)

	router.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(corsHeader)

		r.Get("/", hdlr.index)
		r.Get("/version", hdlr.version)
		r.Get("/favicon.png", hdlr.favicon)
		r.Get("/races", hdlr.races)
		r.Get("/races/{id}", hdlr.race)
		r.Get("/races/{id}/points", hdlr.points)
		r.Get("/metadata/tracks", hdlr.tracksMetadata)
		r.NotFound(notFound)
		r.MethodNotAllowed(notAllowed)
	})

	return router
}

type Handler struct {
	db       *storage.Store
	revision string
}

func (h Handler) pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("pong"))
}

func (h Handler) index(w http.ResponseWriter, r *http.Request) {
	Render(w, r, indexRenderer{})
}

type indexRenderer struct {
	Renderer `json:"-"`
}

func (rd indexRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	return RenderTemplate(r, "index.html", nil)
}

func (h Handler) version(w http.ResponseWriter, r *http.Request) {
	Render(w, r, versionRenderer{Version: h.revision, TemplateData: NewTemplateData(r)})
}

type versionRenderer struct {
	TemplateData `json:"-"`
	Renderer     `json:"-"`
	Version      string `json:"version"`
}

func (rd versionRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	return RenderTemplate(r, "version.html", rd)
}

func (h Handler) favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFileFS(w, r, StaticFS, "static/favicon.png")
}

func corsHeader(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	Render(w, r, NewErrorRenderer(404, "not found", nil, nil))
}

func notAllowed(w http.ResponseWriter, r *http.Request) {
	Render(w, r, NewErrorRenderer(405, "not allowed", nil, nil))
}
