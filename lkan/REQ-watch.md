# lkan — `serve --watch` requirement

**Purpose.** Auto-reload the board when its YAML file changes on disk, without
restarting `lkan serve` or asking the user to refresh the browser.

## Motivation

External tools (e.g. `gh tops board --repo NAME > board.yaml`) regenerate `board.yaml`
periodically. Today users must stop the server, re-run it, and reload the browser tab.
`--watch` makes the board live with respect to file changes.

## Behavior

- Flag: `lkan serve --watch` (defaults to off; opt-in).
- When `--watch` is active, `lkan` monitors the file passed via `-f`
  (default `board.yaml`) for changes.
- On change detection:
  1. Re-parse the YAML.
  2. If parse fails, keep the previous in-memory board and log the error to stderr
     (do **not** crash the server, do **not** clear the UI).
  3. If parse succeeds, replace the in-memory board atomically.
  4. Push a notification to all connected browser clients so they reload (see
     "Browser refresh" below).

## File-change detection

- Use `fsnotify` (`github.com/fsnotify/fsnotify`) for cross-platform file events.
- Watch the file's **directory**, filter events to the target file's basename.
  (Watching the file directly misses atomic-rename writes like `mv` from a temp file,
  which most tools — including `gh tops board > board.yaml` via shell redirection — use.)
- Debounce: coalesce events fired within ~200 ms so a single save doesn't trigger
  multiple reloads.
- Handle removals gracefully: if the file is removed and then re-created
  (atomic-rename pattern), re-add the watch and treat it as a normal change event.

## Browser refresh

Choice of mechanism (pick one; both are acceptable):

- **(a)** Server-Sent Events (SSE) endpoint, e.g. `GET /events`. The page subscribes
  on load; on a board change the server sends `event: reload` and the client calls
  `location.reload()`.
- **(b)** WebSocket on `/ws`. Same shape; sends `{"type":"reload"}`.

SSE is simpler if no other realtime features are planned; WebSocket is more flexible
if `lkan` later wants live mutation push.

Either way, **the client only listens; it does not poll**.

## Non-goals (v1)

- Watching multiple files
- Two-way sync (UI edits writing back to disk — that's a separate feature)
- Conflict resolution between UI state and file state (the UI is read-only on reload —
  the loaded YAML wins)
- Filesystem-level locking

## CLI summary

```
lkan serve                 # current behavior, no watch
lkan serve --watch         # watch -f for changes, auto-reload browser
lkan serve -f board.yaml --watch -http 127.0.0.1:8080
```

## Testing

- Unit: simulate fsnotify events with a temp file; assert the board re-parses.
- Integration: `lkan serve --watch &`, `cp newboard.yaml board.yaml` via atomic rename,
  assert SSE/WS notification is sent.
- Manual: open browser, run `gh tops board --repo release-toptest > board.yaml`,
  verify the UI updates without manual reload.

## Out-of-scope context (informational)

The expected upstream of `board.yaml` is `gh tops board --repo NAME` (a planned
gh CLI extension subcommand from `PicaPD/gh-tops` that exports a Skyhawk-governance
repo's GitHub Issues as `board.yaml`). `lkan` does NOT need to know about that — it
only needs to render whatever `board.yaml` it's pointed at, and reload when the file
changes.
