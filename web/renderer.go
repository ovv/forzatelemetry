package web

import (
	"bytes"
	"log/slog"
	"net/http"
	"reflect"

	"database/sql"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptrace/bun/driver/pgdriver"
)

func init() {
	render.Respond = respond
}

type HTMLRenderer interface {
	render.Renderer
	HTML(http.ResponseWriter, *http.Request) string
}

type Renderer struct{}

func (rd Renderer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func respond(w http.ResponseWriter, r *http.Request, v interface{}) {
	// Format response based on request Accept header.
	switch render.GetAcceptedContentType(r) {
	case render.ContentTypeJSON:
		render.JSON(w, r, v)
	case render.ContentTypeHTML:
		render.HTML(w, r, v.(HTMLRenderer).HTML(w, r))
	default:
		render.JSON(w, r, v)
	}
}

func Render(w http.ResponseWriter, r *http.Request, v HTMLRenderer) {
	err := render.Render(w, r, v)
	if err != nil {
		slog.Error("Renderer failed with error", "error", err.Error(), "type", reflect.TypeOf(err))
		err = render.Render(w, r, NewErrorRenderer(http.StatusInternalServerError, "internal error", err, nil))
		if err != nil {
			slog.Error("Internal error renderer failed with error", "error", err.Error(), "type", reflect.TypeOf(err))
		}
	}
}

func RenderTemplate(r *http.Request, tmpl string, data any) string {
	template, ok := TMPL[tmpl]
	if !ok {
		slog.Error("template not found", "template", tmpl)
		render.Status(r, http.StatusInternalServerError)
		return "internal error"
	}

	var buf bytes.Buffer
	err := template.ExecuteTemplate(&buf, tmpl, data)
	if err != nil {
		slog.Error("error rendering template", "error", err.Error(), "type", reflect.TypeOf(err))
		render.Status(r, http.StatusInternalServerError)
		return "internal error"
	}
	return buf.String()
}

type ErrorRenderer struct {
	TemplateData

	Status  int            `json:"-"`
	Msg     string         `json:"error"`
	Details map[string]any `json:"details"`
	err     error          `json:"-"`
}

func NewErrorRenderer(status int, msg string, err error, details map[string]any) *ErrorRenderer {
	if details == nil {
		details = make(map[string]any)
	}

	return &ErrorRenderer{
		Status:  status,
		Msg:     msg,
		err:     err,
		Details: details,
	}
}

func (rd *ErrorRenderer) Render(w http.ResponseWriter, r *http.Request) error {
	if rd.Status == http.StatusInternalServerError {
		slog.Error("Request failed with unhandled error", "error", rd.err.Error(), "type", reflect.TypeOf(rd.err))
	}
	render.Status(r, rd.Status)
	return nil
}

type ErrorTemplateData struct {
	TemplateData
	*ErrorRenderer
}

func (rd *ErrorRenderer) HTML(w http.ResponseWriter, r *http.Request) string {
	return RenderTemplate(r, "error.html", ErrorTemplateData{
		TemplateData:  NewTemplateData(r),
		ErrorRenderer: rd,
	})
}

func StorageErrorRenderer(err error) *ErrorRenderer {
	switch e := err.(type) {
	case pgdriver.Error:
		msg := e.Field('M')
		status := http.StatusBadRequest
		if msg == "" {
			msg = "internal error"
			status = http.StatusInternalServerError
		}
		return NewErrorRenderer(status, msg, err, nil)
	case *pgconn.PgError:
		msg := "internal error"
		status := http.StatusInternalServerError
		if e.Code == "22P02" {
			status = http.StatusBadRequest
			msg = e.Message
		}
		return NewErrorRenderer(status, msg, err, nil)

	default:
		if err == sql.ErrNoRows {
			return NewErrorRenderer(http.StatusNotFound, "not found", err, nil)
		}
		return NewErrorRenderer(http.StatusInternalServerError, "internal error", err, nil)
	}
}
