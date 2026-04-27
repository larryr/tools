// Package lkan implements a small kanban board with a YAML source of truth
// and a local web UI.
package lkan

// Board is the root document persisted to YAML.
type Board struct {
	Title   string   `yaml:"title"             json:"title"`
	Members []Member `yaml:"members,omitempty" json:"members,omitempty"`
	Labels  []Label  `yaml:"labels,omitempty"  json:"labels,omitempty"`
	Columns []Column `yaml:"columns"           json:"columns"`
}

// Member is a person on the team. Color drives the avatar chip rendered on
// every card they are assigned to.
type Member struct {
	ID    string `yaml:"id"              json:"id"`
	Name  string `yaml:"name"            json:"name"`
	Color string `yaml:"color,omitempty" json:"color,omitempty"`
}

// Label is a reusable tag.
type Label struct {
	Name  string `yaml:"name"            json:"name"`
	Color string `yaml:"color,omitempty" json:"color,omitempty"`
}

// Column is a vertical lane of cards.
type Column struct {
	ID    string `yaml:"id"              json:"id"`
	Title string `yaml:"title"           json:"title"`
	WIP   int    `yaml:"wip,omitempty"   json:"wip,omitempty"`
	Cards []Card `yaml:"cards,omitempty" json:"cards,omitempty"`
}

// Card is a task on the board.
type Card struct {
	ID          string   `yaml:"id"                    json:"id"`
	Title       string   `yaml:"title"                 json:"title"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Assignees   []string `yaml:"assignees,omitempty"   json:"assignees,omitempty"`
	Labels      []string `yaml:"labels,omitempty"      json:"labels,omitempty"`
	Due         string   `yaml:"due,omitempty"         json:"due,omitempty"`
}

// FindColumn returns a pointer to the column with the given id, or nil.
func (b *Board) FindColumn(id string) *Column {
	for i := range b.Columns {
		if b.Columns[i].ID == id {
			return &b.Columns[i]
		}
	}
	return nil
}

// FindCard returns a pointer to the card with the given id, the column it
// lives in, and the index within that column. If not found, returns nil
// pointers and -1.
func (b *Board) FindCard(id string) (*Card, *Column, int) {
	for i := range b.Columns {
		col := &b.Columns[i]
		for j := range col.Cards {
			if col.Cards[j].ID == id {
				return &col.Cards[j], col, j
			}
		}
	}
	return nil, nil, -1
}
