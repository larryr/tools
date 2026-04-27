package mkdoc

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

//go:embed default.css
var defaultCSS string

var htmlTmpl = template.Must(template.New("mkdoc").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="generator" content="mkdoc">
<title>{{.Title}}</title>
<style>
{{.CSS}}
</style>
</head>
<body>
{{.Body}}
</body>
</html>
`))

// renderHTML parses src as Markdown and writes a full standalone HTML
// document to w.
func renderHTML(w io.Writer, src []byte, opts Options) error {
	md := newGoldmark()

	var body bytes.Buffer
	if err := md.Convert(src, &body); err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}

	title := opts.Title
	if title == "" {
		title = firstHeading(src)
	}
	if title == "" {
		title = "document"
	}

	data := struct {
		Title template.HTML
		CSS   template.CSS
		Body  template.HTML
	}{
		Title: template.HTML(html.EscapeString(title)),
		CSS:   template.CSS(defaultCSS),
		Body:  template.HTML(body.String()),
	}
	return htmlTmpl.Execute(w, data)
}

// newGoldmark returns a shared goldmark configuration used by both
// renderers. It enables a GitHub-flavoured-ish subset.
func newGoldmark(extraOpts ...goldmark.Option) goldmark.Markdown {
	opts := []goldmark.Option{
		goldmark.WithExtensions(
			extension.GFM, // tables, strikethrough, linkify, task list
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldhtml.WithUnsafe(), // allow raw HTML in markdown
		),
	}
	opts = append(opts, extraOpts...)
	return goldmark.New(opts...)
}

// firstHeading returns the text of the first ATX heading (# …) in src,
// stripped of the leading hashes and surrounding whitespace. Returns
// "" if no heading is found. Good enough for deriving a document title;
// full AST parsing would be overkill.
var headingRe = regexp.MustCompile(`(?m)^#{1,6}\s+(.+?)\s*#*\s*$`)

func firstHeading(src []byte) string {
	m := headingRe.FindSubmatch(src)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(string(m[1]))
}
