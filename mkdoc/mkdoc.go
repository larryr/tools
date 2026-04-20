// Package mkdoc converts Markdown documents to HTML or PDF.
//
// The public entry point is [Convert], which reads Markdown from an
// io.Reader and writes the rendered output to an io.Writer. Format
// selection and per-format options are carried in [Options].
//
// Both renderers build on github.com/yuin/goldmark for parsing; HTML
// output comes from goldmark's own renderer wrapped in a standalone
// document template, and PDF output is produced by
// github.com/stephenafamo/goldmark-pdf driving a pure-Go gofpdf
// backend. No CGO, no network access, no external binaries.
package mkdoc

import (
	"context"
	"fmt"
	"io"
)

// Format selects the output format for [Convert].
type Format int

const (
	// FormatHTML produces a standalone HTML document.
	FormatHTML Format = iota
	// FormatPDF produces a PDF document.
	FormatPDF
)

// Options controls how Convert renders its input.
//
// Title, when empty, falls back to the first top-level heading in the
// document and, failing that, to "document".
//
// Landscape and PageSize apply only when Format is FormatPDF; they are
// ignored for HTML. PageSize defaults to "A4" when empty. Accepted
// values match gofpdf: A3, A4, A5, Letter, Legal, Tabloid.
type Options struct {
	Format    Format
	Title     string
	Landscape bool
	PageSize  string
}

// Convert reads Markdown from r and writes the rendered output to w
// in the format selected by opts.Format. ctx is honoured by the PDF
// renderer for cancellation; HTML rendering is effectively
// synchronous and ignores it.
func Convert(ctx context.Context, r io.Reader, w io.Writer, opts Options) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	switch opts.Format {
	case FormatHTML:
		return renderHTML(w, src, opts)
	case FormatPDF:
		return renderPDF(ctx, w, src, opts)
	default:
		return fmt.Errorf("unknown format: %d", opts.Format)
	}
}
