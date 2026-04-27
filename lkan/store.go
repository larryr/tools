package lkan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Store wraps an in-memory Board backed by a YAML file. All access is
// serialised through mu. External edits to the file are picked up on the
// next call to maybeReload because each handler stat()s the file's mtime.
type Store struct {
	mu    sync.RWMutex
	path  string
	board *Board
	mtime time.Time
}

// NewStore loads the YAML file at path and returns a Store.
func NewStore(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	var b Board
	if err := yaml.Unmarshal(data, &b); err != nil {
		return fmt.Errorf("parse %s: %w", s.path, err)
	}
	info, err := os.Stat(s.path)
	if err != nil {
		return err
	}
	s.board = &b
	s.mtime = info.ModTime()
	return nil
}

// maybeReload re-reads the YAML file from disk if its mtime has advanced
// since the last load. Caller must hold s.mu (read or write).
func (s *Store) maybeReload() error {
	info, err := os.Stat(s.path)
	if err != nil {
		return err
	}
	if !info.ModTime().After(s.mtime) {
		return nil
	}
	return s.load()
}

// saveAtomic marshals the board, writes to a sibling tmp file, and renames
// it over the source path. Caller must hold s.mu (write).
func (s *Store) saveAtomic() error {
	data, err := yaml.Marshal(s.board)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	tmp, err := os.CreateTemp(dir, ".lkan-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return err
	}
	if info, err := os.Stat(s.path); err == nil {
		s.mtime = info.ModTime()
	}
	return nil
}

// Snapshot returns a deep-copied Board safe for read-only use outside the
// lock. Use this for rendering and JSON responses.
func (s *Store) Snapshot() (*Board, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.maybeReload(); err != nil {
		return nil, err
	}
	return cloneBoard(s.board), nil
}

// Update runs fn under the write lock with a live pointer to the board,
// reloads from disk first, and saves atomically on success.
func (s *Store) Update(fn func(*Board) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.maybeReload(); err != nil {
		return err
	}
	if err := fn(s.board); err != nil {
		return err
	}
	return s.saveAtomic()
}

// Path returns the on-disk path the store is bound to.
func (s *Store) Path() string { return s.path }

func cloneBoard(b *Board) *Board {
	if b == nil {
		return nil
	}
	out := *b
	out.Members = append([]Member(nil), b.Members...)
	out.Labels = append([]Label(nil), b.Labels...)
	out.Columns = make([]Column, len(b.Columns))
	for i, c := range b.Columns {
		col := c
		col.Cards = make([]Card, len(c.Cards))
		for j, card := range c.Cards {
			cd := card
			cd.Assignees = append([]string(nil), card.Assignees...)
			cd.Labels = append([]string(nil), card.Labels...)
			col.Cards[j] = cd
		}
		out.Columns[i] = col
	}
	return &out
}

// ErrNotFound is returned when a card or column id does not exist.
var ErrNotFound = errors.New("not found")
