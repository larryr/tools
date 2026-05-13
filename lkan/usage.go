package lkan

import (
	"bytes"
	"fmt"
	"os"
)

// Delimiters used to mark the lkan-managed section inside an AGENTS.md.
// Everything between (and including) these lines is owned by lkan and may be
// rewritten by `lkan usage --update-agents-md`. Text outside the block is
// preserved verbatim.
const (
	AgentsBeginMarker = "<!-- lkan:begin -->"
	AgentsEndMarker   = "<!-- lkan:end -->"
)

// AgentGuide is the canonical agent-facing documentation for lkan. It is
// printed by `lkan usage` and embedded into AGENTS.md by
// `lkan usage --update-agents-md`. Keep this in sync with the schema in
// board.go.
const AgentGuide = `## lkan — agent guide

` + "`lkan`" + ` is a local kanban board whose source of truth is a single YAML
file (` + "`board.yaml`" + ` by default). The web UI at ` + "`lkan serve`" + ` reads and
writes the same file, so editing the YAML directly is a fully supported way
to drive the board.

### File layout

` + "```yaml" + `
title: Project Name
members:
  - {id: alice, name: Alice, color: "#2E86AB"}
labels:
  - {name: bug,     color: "#d73a4a"}
  - {name: feature, color: "#0e8a16"}
columns:
  - id: todo
    title: To Do
    wip: 3            # optional work-in-progress limit
    cards:
      - id: c-login   # card IDs must be unique across the whole board
        title: "Add login flow"
        description: "Optional longer body."
        assignees: [alice]
        labels: [feature]
        due: "2026-06-01"   # optional, free-form date string
  - id: doing
    title: In Progress
    cards: []
  - id: done
    title: Done
    cards: []
` + "```" + `

### Schema rules

- ` + "`title`" + ` (string, required) — board title shown in the UI header.
- ` + "`members[]`" + ` — each has ` + "`id`" + ` (referenced by ` + "`assignees`" + `), ` + "`name`" + `, optional ` + "`color`" + `.
- ` + "`labels[]`" + ` — each has ` + "`name`" + ` (referenced by ` + "`cards[].labels`" + `) and optional ` + "`color`" + `.
- ` + "`columns[]`" + ` — ordered left-to-right; each has ` + "`id`" + `, ` + "`title`" + `, optional ` + "`wip`" + `, and ` + "`cards[]`" + `.
- ` + "`cards[]`" + ` — ordered top-to-bottom within a column; each has ` + "`id`" + `, ` + "`title`" + `, and optional ` + "`description`" + `, ` + "`assignees`" + `, ` + "`labels`" + `, ` + "`due`" + `.

### Conventions

- Card IDs use the prefix ` + "`c-`" + ` followed by a short slug (` + "`c-login`" + `, ` + "`c-fix-crash`" + `).
  Keep them stable — they appear in commit messages and PR titles.
- Reference a card from a commit or PR with ` + "`[c-id]`" + ` in the subject:
  ` + "`fix(api): tighten input validation [c-login]`" + `. This is the linkage signal
  an automation can rely on to move a card to Done when its commit lands.
- Treat the YAML as the single source of truth. The web UI rewrites it on
  every edit; merge conflicts in this file should be resolved like any other
  source file.

### Common edits

- **Add a card**: append to a column's ` + "`cards`" + ` list with a unique ID.
- **Move a card**: cut the card block from one column's ` + "`cards`" + ` list and
  paste it into another. Position within the list = vertical position.
- **Close a card**: move it to the ` + "`done`" + ` column (do not delete — history
  matters).
- **Rename**: only edit ` + "`title`" + `; never change a card's ` + "`id`" + ` after creation.

### Bootstrapping

- ` + "`lkan init`" + ` writes a starter ` + "`board.yaml`" + ` (refuses to overwrite).
- ` + "`lkan serve`" + ` starts the web UI on ` + "`127.0.0.1:8080`" + ` by default.
- ` + "`lkan usage`" + ` prints this guide.
- ` + "`lkan usage --update-agents-md`" + ` inserts/replaces this guide inside the
  ` + "`<!-- lkan:begin -->…<!-- lkan:end -->`" + ` block in ` + "`./AGENTS.md`" + ` (creating
  the file if it does not exist). Text outside the block is preserved.
`

// AgentBlock returns the AgentGuide wrapped in the lkan:begin/lkan:end
// markers, suitable for splicing into an AGENTS.md.
func AgentBlock() string {
	return AgentsBeginMarker + "\n" + AgentGuide + AgentsEndMarker + "\n"
}

// UpdateAgentsMD inserts or replaces the lkan-managed block inside the file
// at path. If the file does not exist, it is created with just the block.
// If the file exists but contains no lkan block, the block is appended
// (separated by a blank line if the existing file does not end in one).
// If a lkan block is present, its contents are replaced and surrounding
// text is preserved verbatim.
func UpdateAgentsMD(path string) (action string, err error) {
	block := AgentBlock()

	existing, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		if err := os.WriteFile(path, []byte(block), 0o644); err != nil {
			return "", err
		}
		return "created", nil
	}

	beginIdx := bytes.Index(existing, []byte(AgentsBeginMarker))
	endIdx := bytes.Index(existing, []byte(AgentsEndMarker))

	switch {
	case beginIdx >= 0 && endIdx > beginIdx:
		// Replace existing block. endIdx points at start of end marker;
		// include the end marker line itself plus any single trailing
		// newline so the splice is clean.
		tail := endIdx + len(AgentsEndMarker)
		if tail < len(existing) && existing[tail] == '\n' {
			tail++
		}
		var out bytes.Buffer
		out.Write(existing[:beginIdx])
		out.WriteString(block)
		out.Write(existing[tail:])
		if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
			return "", err
		}
		return "updated", nil
	case beginIdx >= 0 || endIdx >= 0:
		return "", fmt.Errorf("%s: found one of %q / %q but not both; refusing to corrupt the file",
			path, AgentsBeginMarker, AgentsEndMarker)
	default:
		var out bytes.Buffer
		out.Write(existing)
		if len(existing) > 0 && !bytes.HasSuffix(existing, []byte("\n\n")) {
			if !bytes.HasSuffix(existing, []byte("\n")) {
				out.WriteByte('\n')
			}
			out.WriteByte('\n')
		}
		out.WriteString(block)
		if err := os.WriteFile(path, out.Bytes(), 0o644); err != nil {
			return "", err
		}
		return "appended", nil
	}
}

