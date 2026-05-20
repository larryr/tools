package lkan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const validYAML = "title: a\ncolumns:\n  - {id: c, title: C}\n"

func TestWatcherFiresOnAtomicRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	if err := os.WriteFile(path, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	wr, err := NewWatcher(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wr.Run(ctx)

	ch, unsub := wr.Subscribe()
	defer unsub()

	// Atomic-rename write: stage the new file under a temp name, then rename
	// it over the target. This mirrors `gh tops board > board.yaml` via shell
	// redirection and editor swap-and-rename saves.
	tmp := filepath.Join(dir, ".staged.tmp")
	if err := os.WriteFile(tmp, []byte("title: b\ncolumns:\n  - {id: c, title: C2}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected reload event after atomic rename, got none")
	}
}

func TestWatcherFiresOnInPlaceWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	if err := os.WriteFile(path, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	wr, err := NewWatcher(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wr.Run(ctx)

	ch, unsub := wr.Subscribe()
	defer unsub()

	if err := os.WriteFile(path, []byte("title: b\ncolumns: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected reload event after in-place write, got none")
	}
}

func TestWatcherIgnoresInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	if err := os.WriteFile(path, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	wr, err := NewWatcher(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wr.Run(ctx)

	ch, unsub := wr.Subscribe()
	defer unsub()

	if err := os.WriteFile(path, []byte("\t- nope: :::\nthis: : is not yaml\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
		t.Fatal("expected no reload event for invalid YAML")
	case <-time.After(debounceWindow + 300*time.Millisecond):
	}
}

func TestWatcherDebouncesBurst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	if err := os.WriteFile(path, []byte(validYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	wr, err := NewWatcher(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wr.Run(ctx)

	ch, unsub := wr.Subscribe()
	defer unsub()

	// Fire 5 quick writes well inside the debounce window. The watcher should
	// emit at most one event for the burst (channel buffer is 1; any further
	// sends drop).
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(path, []byte(validYAML), 0o644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Drain the first event.
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected at least one reload event from burst")
	}

	// No further events should be coming in the next interval.
	select {
	case <-ch:
		t.Fatal("expected burst to coalesce to a single reload event")
	case <-time.After(debounceWindow + 200*time.Millisecond):
	}
}
