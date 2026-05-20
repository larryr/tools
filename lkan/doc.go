/*
lkan is a small kanban board with a YAML source of truth and a local web UI.

Subcommands:

    lkan init     Write a starter board.yaml in the current directory.
    lkan serve    Start the web UI (default if no subcommand).
    lkan usage    Print this document.

Flags:

    -f path       Path to board YAML file       (default: board.yaml)
    -http addr    HTTP listen address           (default: 127.0.0.1:8080)
    -watch        Watch -f for changes and push reload events to open
                  browser tabs over Server-Sent Events (serve only).
                  Default off.

Live reload
-----------

`lkan -watch serve` watches the directory containing board.yaml (so atomic-
rename writes like `gh tops board > board.yaml` are detected) and pushes a
reload event to every open browser tab when the file changes. Invalid YAML
during a write is logged to stderr and ignored — the previous board stays
loaded. Without -watch, none of this runs and behavior is unchanged.

Adopting lkan in an existing repo
---------------------------------

lkan is opt-in and additive. Contributors who don't install it can ignore it.
There is no required label, commit format, CI hook, or GitHub Project. The
only artifact in the repo is board.yaml — commit it or .gitignore it as the
team prefers.

To adopt:

  1. cd into the repo.
  2. lkan init                     # creates a starter board.yaml
  3. Edit board.yaml — add columns, members, labels, and seed cards.
  4. lkan serve                    # http://127.0.0.1:8080

There is no importer from GitHub Issues today (see "Not yet supported" below).
An LLM agent populating board.yaml from existing issues should write the YAML
directly using the schema below.

board.yaml schema
-----------------

Top level:
  title      string         required   Display title for the board.
  members    []Member       optional   People who can be assigned to cards.
  labels     []Label        optional   Reusable card tags.
  columns    []Column       required   Left-to-right kanban columns.

Member:
  id         string         required   Short stable handle (e.g. "me", "alice").
  name       string         required   Display name.
  color      string         optional   CSS hex color, e.g. "#2E86AB".

Label:
  name       string         required   Label text.
  color      string         optional   CSS hex color.

Column:
  id         string         required   Stable column id (e.g. "todo").
  title      string         required   Display title (e.g. "To Do").
  wip        int            optional   WIP limit; 0 or omitted means no limit.
  cards      []Card         optional   Cards in left-to-right / top-to-bottom order.

Card:
  id          string        required   Stable card id; suggested prefix "c-".
  title       string        required   One-line headline.
  description string        optional   Free-form body.
  assignees   []string      optional   Member ids.
  labels      []string      optional   Label names.
  due         string        optional   ISO date "YYYY-MM-DD".

Minimal valid board.yaml
------------------------

    title: My Project Board
    members:
      - {id: me, name: Me, color: "#2E86AB"}
    labels:
      - {name: bug, color: "#d73a4a"}
    columns:
      - id: todo
        title: To Do
        cards:
          - {id: c-1, title: "First card", assignees: [me], labels: [bug]}
      - id: doing
        title: In Progress
        wip: 3
      - id: done
        title: Done

Not yet supported
-----------------

  * No GitHub issue import or sync. There is no `lkan pull`, `lkan reconcile`,
    or `github:` config block in this version. Bridging board.yaml to GitHub
    issues is a design under evaluation — see lkan/NOTES-github-sync.md.
  * No multi-board / multi-project mode. One board.yaml per directory.
  * Moves are persisted by editing board.yaml. There is no separate history.
*/
package lkan
