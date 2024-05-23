package web

import (
	"embed"
)

//go:embed all:html
var TemplateFS embed.FS

//go:embed all:static
var StaticFS embed.FS
