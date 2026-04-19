# mkdoc — Markdown to HTML / PDF / PNG

Status: DRAFT SPEC (for review, not yet implemented)

## Overview

`mkdoc` is a small command-line tool that converts Markdown source files
into one of three output formats:

* **HTML** — CommonMark-compliant HTML fragment or full document.
* **PDF**  — A paginated PDF rendered from the same Markdown.
* **PNG**  — A single rasterised PNG of the rendered document (one image
  per page, or a vertically-stitched single image).

The tool lives alongside the other small CLIs in this repository and
follows the same layout conventions (see `docs/tool-conventions.md`
summarised in the repository `AGENTS.md`).

## Goals

* Pure-Go where possible. The HTML and PDF paths must build with `go
  install` only, no CGO, no external runtime binaries.
* Small surface area: one binary, one flag to pick the output format,
  one flag for the output path. Stdin / stdout friendly.
* Standard library first. Only add a third-party dependency when the
  standard library cannot reasonably do the job.
* Idiomatic Go. Core logic lives in a package (`mkdoc/`), the binary is
  a thin driver (`cmd/mkdoc/`) that wires flags to package calls.
* Deterministic output for a given input + flags (same bytes in → same
  bytes out), so the tool is usable in build pipelines and tests.

## Non-goals

* Full websites, multi-page navigation, sidebars. `prez` already covers
  slideshow/presentation rendering; `mkdoc` is for single documents.
* Authoring features: live preview, watching, auto-reload.
* Theme engines. A single default stylesheet is provided; users who
  want more can pass `-css path/to/custom.css`.
* Editing existing PDFs / PNGs.

## User interface

```
mkdoc [flags] [input.md]

  -f, -format  html|pdf|png   Output format (required unless -o implies it)
  -o, -out     PATH           Output file. "-" or omitted = stdout (HTML only)
  -css         PATH           Optional CSS file to embed (html, pdf, png)
  -title       STRING         Document title (defaults to first H1 or filename)
  -standalone                 Emit a full <html>…</html> doc (html format only)
  -dpi         INT            PNG render DPI (default 96; png only)
  -page-size   STRING         A4|Letter|… (pdf, png; default "A4")
  -v                          Verbose logging to stderr
  -h                          Help
```

Rules:
* If `input.md` is omitted, read Markdown from stdin.
* Format is inferred from the `-o` extension when `-format` is not
  given. `mkdoc -o report.pdf notes.md` is equivalent to `-format pdf`.
* When writing a binary format (pdf, png), `-o` is required — we do
  not dump binary to stdout by default. An explicit `-o -` overrides
  this.
* Unknown flags or a missing required flag exits with status 2 and
  prints usage (matches `flag.Parse` convention).

### Exit codes

* 0 — success
* 1 — runtime error (I/O, rendering, etc.)
* 2 — usage error (bad flags, missing input)

## Examples

```bash
# Markdown → HTML fragment on stdout
mkdoc README.md

# Full standalone HTML document
mkdoc -standalone -o README.html README.md

# PDF with a custom stylesheet
mkdoc -o spec.pdf -css spec.css docs/mkdoc.md

# PNG preview at 144 DPI
mkdoc -o spec.png -dpi 144 docs/mkdoc.md

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
    ├── png.go         # Markdown → PNG
    └── *_test.go
```

Mirrors the existing `gocoap`, `gopuml`, `lcbor`, `prez` layout: logic
in `<tool>/`, entry point in `cmd/<tool>/`.

### Public API sketch

```go
// Package mkdoc converts Markdown documents to HTML, PDF, or PNG.
package mkdoc

type Format int

const (
    FormatHTML Format = iota
    FormatPDF
    FormatPNG
)

type Options struct {
    Format     Format
    Title      string
    CSS        []byte // optional; empty means default stylesheet
    Standalone bool   // HTML only
    PageSize   string // pdf, png; e.g. "A4", "Letter"
    DPI        int    // png only; 0 => 96
}

// Convert reads Markdown from r and writes the rendered output to w.
// It is safe to call from multiple goroutines; no shared mutable state.
func Convert(r io.Reader, w io.Writer, opts Options) error
```

The CLI is a ~60-line driver over `Convert`. That mirrors how `lcbor`,
`modgraphviz`, and `gocoap` already work.

## Dependencies

### HTML — pure Go, already present

Use [`github.com/yuin/goldmark`](https://github.com/yuin/goldmark). It
is already in `go.sum` as an indirect dependency (pulled in by
`golang.org/x/tools`) and is the de-facto Go CommonMark parser — used
by `gopls`, Hugo, and others. MIT licensed. Zero external
dependencies.

Enable extensions that match GitHub-flavoured Markdown: tables, task
lists, strikethrough, auto-heading-IDs. No JavaScript, no remote
stylesheets.

### PDF — pure Go

Use
[`github.com/stephenafamo/goldmark-pdf`](https://github.com/stephenafamo/goldmark-pdf).
It is a `renderer.Renderer` plug-in for goldmark, so the same parse
pipeline handles both HTML and PDF — we swap only the renderer. Pure
Go, MIT licensed, builds without CGO.

Fonts: embed a single permissively-licensed default font set (e.g.
Go's own `golang.org/x/image/font/gofont`) so the binary works offline
with no config. `-css` is accepted for colour/size tweaks but most CSS
properties are ignored — document this limitation in the README.

### PNG — open question, two candidate paths

1. **PDF → PNG via `pdfcpu`** (pure Go). Render PDF first, rasterise
   pages with [`pdfcpu`](https://github.com/pdfcpu/pdfcpu) or
   [`go-pdfium`](https://github.com/klippa-app/go-pdfium). `pdfcpu` is
   pure Go but does not currently do rasterisation; `go-pdfium` uses
   CGO. Neither is ideal.

2. **HTML → PNG via `chromedp`**. High fidelity, but requires a Chrome
   or Chromium binary at runtime — violates the "no external runtime
   binary" goal.

**Proposed resolution for v1**: ship HTML and PDF only. Wire the
`-format png` flag and stub it with a clear `not yet implemented in
this build` error. Revisit once a pure-Go rasteriser is viable, or
accept `chromedp` behind a build tag (`//go:build pngchromedp`) so
the default binary stays lean.

Reviewers: please confirm this trade-off before implementation starts.

## Testing

* Golden-file tests under `mkdoc/testdata/`: pairs of `input.md` and
  expected `output.html`. Regenerate with `go test -run . -update`.
* For PDF, assert structural properties (page count, presence of key
  strings via `pdfcpu` extraction) rather than byte equality — PDFs
  have timestamps.
* PNG tests, when implemented, compare image hashes with a tolerance.
* `go vet ./...` and `go test ./...` must pass before merge. No
  `//nolint` pragmas; fix the underlying issue.

## Build and install

Same as every other tool in this repo:

```bash
go install github.com/larryr/tools/cmd/mkdoc@latest
```

No CGO, no generate step, no embedded web assets beyond the default
font and stylesheet (which are embedded via `//go:embed`).

## Open questions

1. PNG path — pure-Go rasteriser, `chromedp` build tag, or drop PNG
   from v1? (See "PNG — open question" above.)
2. Should `-css` support URL inputs, or local paths only? Local only
   keeps the tool offline and deterministic.
3. Syntax highlighting for fenced code blocks — enable `chroma` (pure
   Go, already transitively available) or leave code blocks plain? Lean
   toward enabling it; small cost, big readability win.
4. Page numbering / footers in PDF — worth a flag in v1, or defer?

## Out of scope for this spec

Implementation details inside each renderer (goldmark extension
configuration, font loading strategy, etc.) are intentionally not
pinned here; they will be decided during implementation and captured
in package-level doc comments.
