# mkdoc — Markdown to HTML / PDF

Status: APPROVED FOR v1 — HTML and PDF only. PNG deferred.

## Overview

`mkdoc` is a small command-line tool that converts a Markdown source
file into one of two output formats:

* **HTML** — a standalone HTML document (one doc, one file), suitable
  for opening directly in a browser.
* **PDF**  — a paginated PDF rendered from the same Markdown, portrait
  by default and landscape on request.

The tool lives alongside the other small CLIs in this repository and
follows the repo-wide layout and style rules in the top-level
[AGENTS.md](../AGENTS.md).

## Goals

* Pure-Go. `go install` only — no CGO, no external runtime binaries,
  no network access at runtime.
* Small surface area: one binary, one flag to pick the output format,
  one flag for the output path, one flag for landscape.
* Standard library first. Only add a third-party dependency when the
  standard library cannot reasonably do the job.
* Idiomatic Go. Core logic lives in `mkdoc/`, the binary is a thin
  driver in `cmd/mkdoc/` that wires flags to package calls.
* Deterministic enough for pipelines: same input + flags → same
  output (minus PDF metadata timestamps).

## Non-goals

* PNG / image output. Deferred — no viable pure-Go rasteriser.
* Full websites, multi-page navigation, sidebars. `prez` already
  handles slideshow/presentation rendering.
* Authoring features: live preview, watching, auto-reload.
* Theme engines. One sensible default is shipped, full stop.

## User interface

```
mkdoc [flags] [input.md]

  -f, -format     html|pdf    Output format (default: inferred from -o,
                              else html)
  -o, -out        PATH        Output file; "-" = stdout (HTML only).
                              Required for PDF unless stdout is explicit.
  -l, -landscape              PDF page orientation: landscape
                              (default: portrait). Ignored for HTML.
  -page-size      SIZE        PDF page size: A4|A3|Letter|Legal
                              (default: A4). Ignored for HTML.
  -title          STRING      Document title. Defaults to the first
                              H1 in the document, falling back to
                              "document" if there is no H1.
  -h                          Help.
```

Rules:
* If `input.md` is omitted, read Markdown from stdin.
* Format is inferred from the `-o` extension when `-format` is not
  given: `.html`/`.htm` → html, `.pdf` → pdf.
* When writing PDF, `-o` is required — we do not dump binary to
  stdout by default. An explicit `-o -` overrides this.
* HTML output is always a complete, standalone document (`<!DOCTYPE
  html>` + `<html>` + `<head>` + `<body>`). There is no "fragment"
  mode.

### Exit codes

* 0 — success
* 1 — runtime error (I/O, rendering, …)
* 2 — usage error (bad flags, missing input)

## Examples

```bash
# Markdown → standalone HTML on stdout
mkdoc README.md > README.html

# Same thing, explicit output file
mkdoc -o README.html README.md

# Portrait PDF (default)
mkdoc -o spec.pdf docs/mkdoc.md

# Landscape PDF
mkdoc -l -o spec.pdf docs/mkdoc.md

# From a pipeline
cat notes.md | mkdoc -format html > notes.html
```

## Architecture

```
tools/
├── cmd/mkdoc/         # main package — flag parsing, wiring, exit codes
│   └── mkdoc.go
└── mkdoc/             # library package — all real work lives here
    ├── mkdoc.go       # public API (Convert, Options, Format)
    ├── html.go        # Markdown → HTML
    ├── pdf.go         # Markdown → PDF
    └── *_test.go
```

Mirrors the existing `gocoap`, `gopuml`, `lcbor`, `prez` layout: logic
in `<tool>/`, entry point in `cmd/<tool>/`.

### Public API sketch

```go
// Package mkdoc converts Markdown documents to HTML or PDF.
package mkdoc

type Format int

const (
    FormatHTML Format = iota
    FormatPDF
)

type Options struct {
    Format    Format
    Title     string
    Landscape bool   // PDF only
    PageSize  string // PDF only; "" → "A4"
}

// Convert reads Markdown from r and writes the rendered output to w.
// The context is used for cancellation during rendering.
func Convert(ctx context.Context, r io.Reader, w io.Writer, opts Options) error
```

The CLI is a ~100-line driver over `Convert`. Same shape as `lcbor`,
`modgraphviz`, and `gocoap`.

## Dependencies

### HTML — pure Go

[`github.com/yuin/goldmark`](https://github.com/yuin/goldmark). Already
transitively present via `golang.org/x/tools`; promoted to a direct
dependency. MIT licensed, no CGO, no extra transitive deps of its own.

Enabled extensions (GitHub-flavoured-ish subset): tables, strikethrough,
task lists, auto heading IDs. No JavaScript, no remote assets.

The HTML wrapper is a tiny embedded `text/template` with a minimal
stylesheet hard-coded in the package (`//go:embed default.css`). No
user-supplied CSS in v1 — keeps the tool single-purpose.

### PDF — pure Go

[`github.com/stephenafamo/goldmark-pdf`](https://github.com/stephenafamo/goldmark-pdf)
v0.4.2. A `renderer.Renderer` plug-in for goldmark, so the same parse
pipeline handles both HTML and PDF — we swap only the renderer.

Transitive chain (all pure Go, MIT/BSD/Apache-style):

* `github.com/phpdave11/gofpdf`         — PDF writer
* `github.com/alecthomas/chroma/v2`     — code-block syntax highlight
* `github.com/go-swiss/fonts`           — font helper (not exercised
  at runtime when we stick to inbuilt fonts)
* `github.com/jellydator/ttlcache/v3`
* `github.com/dlclark/regexp2`
* `golang.org/x/sync`

**Offline by default.** The library can download Google Fonts on the
fly; we avoid this by explicitly configuring the three inbuilt gofpdf
fonts — `FontHelvetica` for body and headings, `FontCourier` for code.
No network access at runtime.

Orientation is set via `pdf.WithFpdf(ctx, pdf.FpdfConfig{Orientation:
"L" or "P", PaperSize: ...})`.

## Testing

* Golden-file tests under `mkdoc/testdata/`: pairs of `input.md` and
  expected `output.html`.
* For PDF, assert structural properties (non-zero bytes, `%PDF-` magic,
  trailer present, page count via `gofpdf` APIs or a minimal parser)
  rather than byte equality — PDFs contain timestamps.
* `go vet ./...` and `go test ./...` must pass before merge.

## Build and install

```bash
go install github.com/larryr/tools/cmd/mkdoc@latest
```

No CGO, no generate step, no embedded web assets beyond the default
stylesheet (embedded via `//go:embed`).

## Future work (not v1)

* PNG output. Requires a pure-Go rasteriser or a `chromedp` build
  tag; revisit when one of those is acceptable.
* User-supplied CSS / theme. Easy to add once we decide we want it.
* Syntax highlighting colours in HTML output — `chroma` is already
  pulled in transitively; could wire it up cheaply.
* Page numbers / footers in PDF.
