package mkdoc

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestConvertHTML(t *testing.T) {
	src := "# Title\n\nHello **world** with `code` and a [link](https://example.com).\n"

	var out bytes.Buffer
	err := Convert(context.Background(), strings.NewReader(src), &out, Options{Format: FormatHTML})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	got := out.String()
	// Spot-check the standalone wrapper and the rendered body.
	for _, want := range []string{
		"<!DOCTYPE html>",
		"<title>Title</title>",
		"<h1 ",
		"Title</h1>",
		"<strong>world</strong>",
		`<code>code</code>`,
		`href="https://example.com"`,
		"</body>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("HTML output missing %q\n---\n%s\n---", want, got)
		}
	}
}

func TestConvertHTMLTitleFallback(t *testing.T) {
	// No H1 in the source: the template should fall back to "document".
	src := "Just a paragraph.\n"
	var out bytes.Buffer
	if err := Convert(context.Background(), strings.NewReader(src), &out, Options{Format: FormatHTML}); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(out.String(), "<title>document</title>") {
		t.Errorf("expected default title 'document'; got:\n%s", out.String())
	}
}

func TestConvertHTMLExplicitTitle(t *testing.T) {
	src := "# Irrelevant Heading\n\nbody\n"
	var out bytes.Buffer
	if err := Convert(context.Background(), strings.NewReader(src), &out,
		Options{Format: FormatHTML, Title: "My Title"}); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if !strings.Contains(out.String(), "<title>My Title</title>") {
		t.Errorf("explicit -title not honoured:\n%s", out.String())
	}
}

func TestConvertPDFPortrait(t *testing.T) {
	src := "# Hello\n\nA paragraph.\n"
	var out bytes.Buffer
	if err := Convert(context.Background(), strings.NewReader(src), &out,
		Options{Format: FormatPDF}); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	assertPDF(t, out.Bytes())
}

func TestConvertPDFLandscape(t *testing.T) {
	src := "# Hello\n\nA paragraph.\n"
	var out bytes.Buffer
	if err := Convert(context.Background(), strings.NewReader(src), &out,
		Options{Format: FormatPDF, Landscape: true}); err != nil {
		t.Fatalf("Convert: %v", err)
	}
	assertPDF(t, out.Bytes())
}

func TestConvertUnknownFormat(t *testing.T) {
	var out bytes.Buffer
	err := Convert(context.Background(), strings.NewReader("x"), &out, Options{Format: Format(99)})
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestFirstHeading(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"# Hello\n", "Hello"},
		{"## Sub\n", "Sub"},
		{"no heading here\n", ""},
		{"# Trailing hashes ##\n", "Trailing hashes"},
		{"text\n\n# Second wins? no, first\n\n# second\n", "Second wins? no, first"},
	}
	for _, c := range cases {
		if got := firstHeading([]byte(c.in)); got != c.want {
			t.Errorf("firstHeading(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// assertPDF does lightweight structural validation on PDF bytes: the
// magic header, a %%EOF trailer, and a non-trivial size. Full parsing
// would pull in heavy deps we don't want in a unit test.
func assertPDF(t *testing.T, b []byte) {
	t.Helper()
	if len(b) < 200 {
		t.Fatalf("PDF too small: %d bytes", len(b))
	}
	if !bytes.HasPrefix(b, []byte("%PDF-")) {
		t.Errorf("missing %%PDF- header; first bytes: %q", b[:min(16, len(b))])
	}
	if !bytes.Contains(b[max(0, len(b)-64):], []byte("%%EOF")) {
		t.Errorf("missing %%%%EOF trailer")
	}
}
