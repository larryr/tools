package lkan

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func writeBoard(t *testing.T, path string, b *Board) {
	t.Helper()
	data, err := yaml.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	writeBoard(t, path, sampleBoard())

	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got := cardIDs(snap.Columns[0]); !eq(got, []string{"a", "b", "c"}) {
		t.Errorf("todo: %v", got)
	}
}

func TestStore_UpdateAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	writeBoard(t, path, sampleBoard())

	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Update(func(b *Board) error {
		return MoveCard(b, "b", "doing", 0)
	}); err != nil {
		t.Fatal(err)
	}

	// Re-read from disk to confirm persistence.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "id: b") {
		t.Fatalf("YAML missing card b:\n%s", data)
	}

	// No leftover tmp file.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".lkan-") {
			t.Errorf("leftover tmp file: %s", e.Name())
		}
	}

	// Reload and verify column membership.
	s2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	snap, _ := s2.Snapshot()
	if got := cardIDs(snap.Columns[1]); !eq(got, []string{"b", "d"}) {
		t.Errorf("doing: %v", got)
	}
}

func TestStore_ExternalEditReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	writeBoard(t, path, sampleBoard())

	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	// Bump mtime far enough to be detectable on coarse-grained filesystems.
	future := time.Now().Add(2 * time.Second)
	b2 := sampleBoard()
	b2.Title = "External"
	writeBoard(t, path, b2)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}

	snap, err := s.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if snap.Title != "External" {
		t.Errorf("title not reloaded: %q", snap.Title)
	}
}

func TestStore_ConcurrentUpdates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	writeBoard(t, path, sampleBoard())

	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = s.Update(func(b *Board) error {
				_, err := AddCard(b, "todo", Card{Title: "x"})
				return err
			})
		}()
	}
	wg.Wait()

	snap, _ := s.Snapshot()
	if got := len(snap.Columns[0].Cards); got != 23 {
		t.Errorf("expected 23 cards in todo, got %d", got)
	}
}
