# CLAUDE.md

This file is read by Claude Code when working in this repository.
Everything it needs is in [AGENTS.md](./AGENTS.md) — treat that as the
authoritative guide for layout, style, and contribution rules.

## Quick reference for Claude

* **Repo shape:** one Go module, many small CLIs. Each tool has a
  library package at `<toolname>/` and a thin driver at
  `cmd/<toolname>/`. Copy this shape for new tools.
* **Style:** simplified Google Go style — stdlib first, no CGO, no
  vendored deps, `gofmt` / `go vet` clean, `io.Reader` / `io.Writer`
  for core funcs, `context.Context` for anything long-running.
* **Specs before code:** new tools land a short design doc in
  `docs/<toolname>.md` first. See `docs/mkdoc.md` for the template.
* **Branches:** work on the branch assigned for the task. Do not
  push elsewhere. Do not open a PR unless explicitly asked.
* **Dependencies:** justify every new `require` in the spec or PR
  description. Prefer the standard library.

Read [AGENTS.md](./AGENTS.md) in full before the first edit in a new
session.
