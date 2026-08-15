package prez

import (
	"bytes"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// mdTemplate wraps rendered markdown in a page shell matching dir.tmpl.
// It is embedded rather than loaded from basePath so markdown rendering
// works with any -base directory, including the x/tools default.
var mdTemplate = template.Must(template.New("markdown").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
  <title>{{.Title}}</title>
  <link type="text/css" rel="stylesheet" href="/static/dir.css">
</head>
<body>

<div id="topbar"><div class="container">
<div id="heading"><a href="/">Presentations</a></div>
</div></div>

<div id="page">
{{.Body}}
</div>

</body>
</html>
`))

// Raw HTML in the source is escaped by goldmark's default renderer.
var mdRenderer = goldmark.New(goldmark.WithExtensions(extension.GFM))

func isMarkdown(path string) bool {
	return filepath.Ext(path) == ".md"
}

// renderMarkdown converts the markdown file to HTML and writes it to w
// wrapped in mdTemplate.
func renderMarkdown(w io.Writer, docFile string) error {
	src, err := os.ReadFile(docFile)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := mdRenderer.Convert(src, &buf); err != nil {
		return err
	}
	return mdTemplate.Execute(w, struct {
		Title string
		Body  template.HTML
	}{
		Title: filepath.Base(docFile),
		Body:  template.HTML(buf.String()),
	})
}
