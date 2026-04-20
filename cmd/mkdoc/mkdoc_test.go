package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larryr/tools/mkdoc"
)

func TestResolveFormat(t *testing.T) {
	cases := []struct {
		flag, out string
		want      mkdoc.Format
		wantErr   bool
	}{
		{"", "", mkdoc.FormatHTML, false},
		{"html", "", mkdoc.FormatHTML, false},
		{"HTML", "", mkdoc.FormatHTML, false},
		{"pdf", "", mkdoc.FormatPDF, false},
		{"", "out.pdf", mkdoc.FormatPDF, false},
		{"", "OUT.HTM", mkdoc.FormatHTML, false},
		{"", "out.html", mkdoc.FormatHTML, false},
		{"", "-", mkdoc.FormatHTML, false},
		{"html", "out.pdf", mkdoc.FormatHTML, false}, // explicit flag wins
		{"png", "", 0, true},
	}
	for _, c := range cases {
		got, err := resolveFormat(c.flag, c.out)
		if (err != nil) != c.wantErr {
			t.Errorf("resolveFormat(%q,%q) err=%v, wantErr=%v", c.flag, c.out, err, c.wantErr)
			continue
		}
		if err == nil && got != c.want {
			t.Errorf("resolveFormat(%q,%q) = %v, want %v", c.flag, c.out, got, c.want)
		}
	}
}

func TestParseFlags(t *testing.T) {
	f, err := parseFlags([]string{"-l", "-f", "pdf", "-o", "out.pdf", "-title", "T", "in.md"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if !f.landscape || f.format != "pdf" || f.out != "out.pdf" || f.title != "T" || f.input != "in.md" {
		t.Errorf("parseFlags mismatch: %+v", f)
	}
}

func TestParseFlagsUnknown(t *testing.T) {
	var buf bytes.Buffer
	_, err := parseFlags([]string{"-xyz"}, &buf)
	var u *usageError
	if !errors.As(err, &u) {
		t.Errorf("expected usageError; got %v", err)
	}
}

func TestOpenOutputPDFRequiresFile(t *testing.T) {
	_, _, err := openOutput("", mkdoc.FormatPDF, &bytes.Buffer{})
	var u *usageError
	if !errors.As(err, &u) {
		t.Errorf("expected usageError for PDF + missing -o; got %v", err)
	}
}

func TestRunHTMLStdinStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run(nil, strings.NewReader("# Hi\n\nbody\n"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout.String(), "<title>Hi</title>") {
		t.Errorf("expected stdin-sourced H1 to become title; got:\n%s", stdout.String())
	}
}

func TestRunPDFFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.md")
	out := filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(in, []byte("# Hi\n\nbody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if err := run([]string{"-o", out, in}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("run: %v", err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF-")) {
		t.Errorf("expected PDF header; got %q", b[:16])
	}
}
