package lkan

import (
	"fmt"
	"os"
)

// starterYAML is the contents written by `lkan init`.
const starterYAML = `title: My Project Board
members:
  - {id: me, name: Me, color: "#2E86AB"}
labels:
  - {name: bug,     color: "#d73a4a"}
  - {name: feature, color: "#0e8a16"}
  - {name: chore,   color: "#a2a2a2"}
columns:
  - id: todo
    title: To Do
    cards:
      - {id: c-welcome, title: "Welcome to lkan! Drag me into In Progress.", assignees: [me], labels: [feature]}
      - {id: c-edit,    title: "Edit board.yaml directly to add team and cards.", labels: [chore]}
  - id: doing
    title: In Progress
    wip: 3
    cards: []
  - id: done
    title: Done
    cards: []
`

// WriteStarter creates path with the starter board contents. It refuses to
// overwrite an existing file.
func WriteStarter(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("init %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(starterYAML); err != nil {
		return err
	}
	return nil
}
