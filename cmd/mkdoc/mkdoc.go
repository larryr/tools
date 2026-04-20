// Command mkdoc converts a Markdown file to HTML or PDF.
//
// Usage:
//
//	mkdoc [flags] [input.md]
//
// With no input file, mkdoc reads from standard input. See -h for the
// full flag list.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/larryr/tools/mkdoc"
)

const usageText = `Usage: mkdoc [flags] [input.md]

Convert a Markdown file to HTML or PDF. With no input file, reads stdin.

Flags:
  -f, -format FORMAT   Output format: html or pdf.
                       If omitted, inferred from -o's extension, else html.
  -o, -out PATH        Output file. "-" writes to stdout (HTML only).
                       Required for PDF unless "-o -" is given explicitly.
  -l, -landscape       PDF page orientation: landscape (default: portrait).
                       Ignored for HTML.
  -page-size SIZE      PDF page size: A3, A4, A5, Letter, Legal, Tabloid
                       (default: A4). Ignored for HTML.
  -title STRING        Document title. Defaults to the first H1 in the
                       document, else "document".
  -h                   Show this help.

Examples:
  mkdoc README.md                         # standalone HTML to stdout
  mkdoc -o README.html README.md          # standalone HTML to file
  mkdoc -o spec.pdf docs/mkdoc.md         # portrait PDF
  mkdoc -l -o spec.pdf docs/mkdoc.md      # landscape PDF
`

// flags holds the parsed command-line configuration. A struct (rather
// than package-level vars) keeps main testable and avoids the usual
// flag-package global state.
type flags struct {
	format    string
	out       string
	landscape bool
	pageSize  string
	title     string
	input     string // positional; "" means stdin
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var usageErr *usageError
		if errors.As(err, &usageErr) {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprint(os.Stderr, usageText)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "mkdoc:", err)
		os.Exit(1)
	}
}

type usageError struct{ msg string }

func (e *usageError) Error() string { return "mkdoc: " + e.msg }

func usagef(f string, a ...any) error { return &usageError{msg: fmt.Sprintf(f, a...)} }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	f, err := parseFlags(args, stderr)
	if err != nil {
		return err
	}

	format, err := resolveFormat(f.format, f.out)
	if err != nil {
		return err
	}

	in, closeIn, err := openInput(f.input, stdin)
	if err != nil {
		return err
	}
	defer closeIn()

	out, closeOut, err := openOutput(f.out, format, stdout)
	if err != nil {
		return err
	}
	defer closeOut()

	opts := mkdoc.Options{
		Format:    format,
		Title:     f.title,
		Landscape: f.landscape,
		PageSize:  f.pageSize,
	}

	return mkdoc.Convert(context.Background(), in, out, opts)
}

// parseFlags reads argv-style args into a flags struct. It supports
// the short/long pairs (-f/-format, -o/-out, -l/-landscape) without
// pulling in a third-party CLI library.
func parseFlags(args []string, stderr io.Writer) (*flags, error) {
	f := &flags{pageSize: "A4"}

	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "-h", "-help", "--help":
			fmt.Fprint(stderr, usageText)
			os.Exit(0)
		case "-l", "-landscape", "--landscape":
			f.landscape = true
		case "-f", "-format", "--format":
			v, err := takeValue(args, &i, a)
			if err != nil {
				return nil, err
			}
			f.format = v
		case "-o", "-out", "--out":
			v, err := takeValue(args, &i, a)
			if err != nil {
				return nil, err
			}
			f.out = v
		case "-page-size", "--page-size":
			v, err := takeValue(args, &i, a)
			if err != nil {
				return nil, err
			}
			f.pageSize = v
		case "-title", "--title":
			v, err := takeValue(args, &i, a)
			if err != nil {
				return nil, err
			}
			f.title = v
		case "--":
			i++
			if i < len(args) {
				f.input = args[i]
			}
			i = len(args)
			continue
		default:
			if strings.HasPrefix(a, "-") && a != "-" {
				return nil, usagef("unknown flag: %s", a)
			}
			if f.input != "" {
				return nil, usagef("unexpected argument: %s", a)
			}
			f.input = a
		}
		i++
	}
	return f, nil
}

func takeValue(args []string, i *int, flag string) (string, error) {
	if *i+1 >= len(args) {
		return "", usagef("flag %s requires a value", flag)
	}
	*i++
	return args[*i], nil
}

// resolveFormat resolves the final output format from the -format flag
// and the -o path. An explicit flag wins; otherwise we sniff the
// extension; otherwise HTML.
func resolveFormat(flag, out string) (mkdoc.Format, error) {
	v := strings.ToLower(strings.TrimSpace(flag))
	if v == "" && out != "" && out != "-" {
		switch strings.ToLower(filepath.Ext(out)) {
		case ".html", ".htm":
			v = "html"
		case ".pdf":
			v = "pdf"
		}
	}
	switch v {
	case "", "html":
		return mkdoc.FormatHTML, nil
	case "pdf":
		return mkdoc.FormatPDF, nil
	default:
		return 0, usagef("unknown format %q (want html or pdf)", flag)
	}
}

func openInput(path string, stdin io.Reader) (io.Reader, func(), error) {
	if path == "" {
		return stdin, func() {}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("opening input: %w", err)
	}
	return file, func() { _ = file.Close() }, nil
}

func openOutput(path string, format mkdoc.Format, stdout io.Writer) (io.Writer, func(), error) {
	switch path {
	case "":
		if format == mkdoc.FormatPDF {
			return nil, nil, usagef("-o is required for PDF output (use '-o -' to force stdout)")
		}
		return stdout, func() {}, nil
	case "-":
		return stdout, func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("creating output: %w", err)
	}
	return file, func() { _ = file.Close() }, nil
}
