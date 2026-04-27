package lkan

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

// boardTmpl is the parsed template for the main board page.
var boardTmpl = template.Must(template.ParseFS(templatesFS, "templates/board.html.tmpl"))

// StaticFS returns the embedded /static tree rooted so that
// http.FileServer serves /static/kanban.js etc.
func StaticFS() fs.FS {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return sub
}
