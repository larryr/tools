package mkdoc

import (
	"context"
	"fmt"
	"io"

	pdf "github.com/stephenafamo/goldmark-pdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

// renderPDF parses src as Markdown and writes a PDF document to w.
//
// The PDF uses gofpdf's three built-in fonts (Helvetica / Courier) so
// that rendering is fully offline — goldmark-pdf's default Google Font
// backend is not exercised.
func renderPDF(ctx context.Context, w io.Writer, src []byte, opts Options) error {
	orientation := "P"
	if opts.Landscape {
		orientation = "L"
	}
	paper := opts.PageSize
	if paper == "" {
		paper = "A4"
	}
	title := opts.Title
	if title == "" {
		title = firstHeading(src)
	}
	if title == "" {
		title = "document"
	}

	fpdfCfg := pdf.FpdfConfig{
		Title:       title,
		Orientation: orientation,
		PaperSize:   paper,
	}

	renderer := pdf.New(
		pdf.WithContext(ctx),
		pdf.WithFpdf(ctx, fpdfCfg),
		pdf.WithHeadingFont(pdf.FontHelvetica),
		pdf.WithBodyFont(pdf.FontHelvetica),
		pdf.WithCodeFont(pdf.FontCourier),
	)

	md := goldmark.New(
		goldmark.WithRenderer(renderer),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)

	if err := md.Convert(src, w); err != nil {
		return fmt.Errorf("rendering pdf: %w", err)
	}
	return nil
}
