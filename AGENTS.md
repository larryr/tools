# AGENTS.md — how to work in this repo

This file gives coding agents (and humans) the ground rules for
contributing tools to `github.com/larryr/tools`. It is intentionally
short — read it all before writing code.

## What this repo is

A single Go module containing a handful of small, independent command
line tools used during development (e.g. `gopuml`, `gocoap`, `lcbor`,
`modgraphviz`, `prez`, `ldhcpcli`, `mkdoc`). Each tool is its own
island: no tool depends on another, and there is no shared framework.

## Layout convention

Every tool follows the same two-directory shape:

```
<toolname>/           # library package — all real work lives here
    <toolname>.go     # public API
    *.go              # helpers, split by topic
    *_test.go         # tests next to the code they cover
cmd/<toolname>/       # main package — a thin CLI driver
    <toolname>.go     # flag parsing, wiring, exit codes
```

The driver in `cmd/<toolname>/` should be small — ideally under ~100
lines. It parses flags, opens files, and hands `io.Reader` /
`io.Writer` to the library package. All logic, all testability, all
reuse lives in the library package.

New tool? Copy the shape of `gocoap` (non-trivial library + thin CLI)
or `modgraphviz` (tiny, self-contained). Do not invent a new layout.

## Go style — simplified Google style

We follow the [Go style guide](https://google.github.io/styleguide/go/)
with the following simplifications for a single-maintainer repo:

* **Standard library first.** Reach for `flag`, `io`, `bufio`,
  `encoding/*`, `net/http`, `text/template` before any third-party
  package. Only add a dependency when the stdlib genuinely can't do the
  job, and justify it in the PR / spec.
* **No CGO** unless a tool truly needs it. Pure-Go builds only.
* **`gofmt` / `goimports` clean.** No exceptions.
* **`go vet ./...` must pass** before commit.
* **Errors:** return them, wrap with `fmt.Errorf("doing X: %w", err)`,
  don't `log.Fatal` from library code. Only `main` may exit.
* **I/O:** core functions take `io.Reader` / `io.Writer`. Do not bake
  in `os.Stdin` / `os.Stdout` below the driver.
* **Context:** any function doing network, file, or long-running work
  accepts `context.Context` as the first argument.
* **Naming:** packages are short, lowercase, no underscores. Exported
  identifiers have doc comments that start with the identifier name.
* **Comments:** document the *why* when it isn't obvious. Skip
  comments that merely restate the code. Package-level doc comments on
  the main `.go` file are required.
* **Tests:** table-driven where practical. Golden files live under
  `testdata/`. `go test ./...` must pass before commit.

## What NOT to do

* Don't add a framework, a plugin system, or a shared `internal/`
  package "for future tools". YAGNI. If three tools actually need the
  same helper, factor it then.
* Don't introduce config files, env vars, or dotfiles. Tools are
  driven by flags and stdin.
* Don't add emojis, colour output, or spinners unless the tool is
  explicitly interactive.
* Don't vendor. Go modules only.
* Don't silently depend on external binaries (`dot`, `chrome`,
  `pandoc`). If a tool genuinely needs one, document it in its README
  and fail with a clear error message when it's missing.

## Adding a new tool

1. Write a short spec in `docs/<toolname>.md` describing the user
   interface, dependencies, and trade-offs. Get it reviewed before
   writing code — see `docs/mkdoc.md` for the template.
2. Create the library package at `<toolname>/` with a
   `Convert`-style or `New*`-style public API.
3. Create the driver at `cmd/<toolname>/<toolname>.go`.
4. Add a one-line entry under the tool list in the top-level
   `README.md`.
5. `go install ./cmd/<toolname>` must work.

## Branches and commits

* Feature branches are named `claude/<short-description>-<nonce>`
  when opened by the coding agent.
* Commit messages follow the repo convention visible in
  `git log`: short imperative subject, blank line, optional body. No
  `Co-Authored-By` lines.
* Do not push to `main` or any branch other than the one assigned for
  the task.
* Do not open a pull request unless explicitly asked.

## Agent-specific guidance

* Prefer reading existing tools (`gocoap/`, `gopuml/`, `lcbor/`,
  `modgraphviz.go`) before designing a new one — they are the ground
  truth for "what we do here" and will always be more current than
  this document.
* When asked for a spec, write a spec — don't silently implement. Land
  the spec in `docs/`, commit, and stop.
* When implementing, keep diffs minimal. Don't refactor neighbouring
  tools for consistency in the same change.
