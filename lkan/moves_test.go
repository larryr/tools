package lkan

import (
	"errors"
	"testing"
)

func sampleBoard() *Board {
	return &Board{
		Title: "T",
		Columns: []Column{
			{ID: "todo", Title: "To Do", Cards: []Card{
				{ID: "a", Title: "A"},
				{ID: "b", Title: "B"},
				{ID: "c", Title: "C"},
			}},
			{ID: "doing", Title: "Doing", Cards: []Card{
				{ID: "d", Title: "D"},
			}},
			{ID: "done", Title: "Done"},
		},
	}
}

func cardIDs(col Column) []string {
	out := make([]string, len(col.Cards))
	for i, c := range col.Cards {
		out[i] = c.ID
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMoveCard_AcrossColumns(t *testing.T) {
	b := sampleBoard()
	if err := MoveCard(b, "b", "doing", 0); err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(b.Columns[0]); !eq(got, []string{"a", "c"}) {
		t.Errorf("todo: %v", got)
	}
	if got := cardIDs(b.Columns[1]); !eq(got, []string{"b", "d"}) {
		t.Errorf("doing: %v", got)
	}
}

func TestMoveCard_WithinColumn(t *testing.T) {
	b := sampleBoard()
	// a, b, c -> move c to index 0 -> c, a, b
	if err := MoveCard(b, "c", "todo", 0); err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(b.Columns[0]); !eq(got, []string{"c", "a", "b"}) {
		t.Errorf("todo: %v", got)
	}
}

func TestMoveCard_IndexClamps(t *testing.T) {
	b := sampleBoard()
	if err := MoveCard(b, "a", "done", 99); err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(b.Columns[2]); !eq(got, []string{"a"}) {
		t.Errorf("done: %v", got)
	}
	if err := MoveCard(b, "a", "todo", -5); err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(b.Columns[0]); !eq(got, []string{"a", "b", "c"}) {
		t.Errorf("todo: %v", got)
	}
}

func TestMoveCard_UnknownCard(t *testing.T) {
	b := sampleBoard()
	err := MoveCard(b, "nope", "doing", 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestMoveCard_UnknownColumn(t *testing.T) {
	b := sampleBoard()
	err := MoveCard(b, "a", "nope", 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestAddCard(t *testing.T) {
	b := sampleBoard()
	id, err := AddCard(b, "doing", Card{Title: "new"})
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected generated id")
	}
	if got := cardIDs(b.Columns[1]); !eq(got, []string{"d", id}) {
		t.Errorf("doing: %v", got)
	}
}

func TestAddCard_RequiresTitle(t *testing.T) {
	b := sampleBoard()
	if _, err := AddCard(b, "todo", Card{Title: "  "}); err == nil {
		t.Fatal("expected error on empty title")
	}
}

func TestEditCard(t *testing.T) {
	b := sampleBoard()
	title := "renamed"
	if err := EditCard(b, "a", CardPatch{Title: &title}); err != nil {
		t.Fatal(err)
	}
	c, _, _ := b.FindCard("a")
	if c.Title != "renamed" {
		t.Errorf("title: %q", c.Title)
	}
}

func TestDeleteCard(t *testing.T) {
	b := sampleBoard()
	if err := DeleteCard(b, "b"); err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(b.Columns[0]); !eq(got, []string{"a", "c"}) {
		t.Errorf("todo: %v", got)
	}
}
