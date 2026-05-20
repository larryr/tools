# lkan GitHub sync — design notes (proposal)

**Status:** Proposal under evaluation. None of the commands, config keys, or
behaviors below are implemented in the current `lkan` binary. Do not treat
this file as user documentation — see `lkan usage` for what actually works.

Audience: project owners (and the author) considering whether to extend lkan
with GitHub-issue awareness, and what shape that extension should take.

## Posture

lkan is **opt-in and additive**. Any GitHub integration should preserve that:
no required labels, no required commit-message format, no required CI hook,
no GitHub Project. A contributor who never installs lkan should still be able
to participate normally; one who runs `lkan serve` should still get a useful
local board.

The proposal: gate operating mode on a single `github:` block at the top of
`board.yaml`. Omit the block and lkan stays purely local (today's behavior).
Add it and lkan becomes aware of the repo's issues to the degree the owner
configures.

## Three proposed operating modes

| Mode | `github:` block | Behavior |
|---|---|---|
| **Local-only** | absent | `board.yaml` is the source of truth. No network. |
| **Mirror** (read-only) | present, `mode: pull` | `lkan pull` reads issues; local moves stay local. |
| **Reconcile** (bidirectional) | present, `mode: reconcile` | `lkan reconcile` applies local moves back to issues. |

Owners would promote upward as trust grows: start local, add pull, add reconcile.

## Optional conventions

Adopt none, some, or all. Each independently useful.

### Status labels (column mapping)

Beyond `open` (→ Backlog) and `closed` (→ Done), lkan could place cards in
intermediate columns based on labels chosen by the owner. Suggested defaults:

- `status:in-progress` → In Progress column
- `status:review` → Review column

Names configurable in `board.yaml`. Rename them, or skip them and accept a
2-column board.

### Scoping label

For large repos, a single label (e.g. `board`, or `lkan`) would mark which
issues belong on the kanban. lkan filters its pull by that label. Without it,
lkan would pull every open issue.

### Card-ID conventions

- `gh-N` — card mirrors GitHub issue #N. Created automatically on `lkan pull`.
- `c-<slug>` — local-only card (an idea not yet committed to an issue). When
  such a card moves past the column named in `github.promote_from`,
  `lkan reconcile` would open an issue and rename the card.

### Commit-message linkage

Putting `[gh-N]` (or `[c-id]`) in a commit subject would let reconcile relate
commits to cards/issues without guessing. A convention, not a requirement —
reconcile would fall back to title/state matching when absent.

## Deliberate non-goals (even if the proposal lands)

- Do not auto-rename GitHub issue titles when a card is renamed.
- Do not auto-close on a card move without an explicit `--apply` step.
- Do not require a GitHub Project — issues alone should be sufficient.
- Do not mutate `AGENTS.md` outside an explicit `<!-- lkan:begin -->` block,
  and never on `lkan init`.

## Proposed adoption path (if implemented)

1. `lkan init` — start local-only. Commit `board.yaml`, or `.gitignore` it,
   per team preference.
2. Live with it for a sprint. If the team gravitates to the visual surface,
   add a `github:` block in `pull` mode.
3. Once the mapping looks right after a few `lkan pull` runs, switch to
   `reconcile` mode and run `lkan reconcile` (dry-run) before any `--apply`.
4. Only add a scoping label if pull volume becomes a nuisance.

## Possible CONTRIBUTING.md snippet (if shipped)

```markdown
### Optional: lkan for a local kanban view

This project's issues can be viewed as a kanban board with
[lkan](https://github.com/larryr/tools/tree/main/lkan). Installation and
use are entirely optional — contributing via the GitHub issues tab works
unchanged.

If you want the board view:

    go install github.com/larryr/tools/cmd/lkan@latest
    lkan pull        # populates board.yaml from this repo's issues
    lkan serve       # opens the board at http://127.0.0.1:8080

Moves you make in the board stay local until you run `lkan reconcile`.
```
