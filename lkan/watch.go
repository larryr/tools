package lkan

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// debounceWindow coalesces rapid filesystem events. Atomic-rename writes
// (e.g. `gh tops board > board.yaml`, editor swap-and-rename) typically
// fire 2-3 events within a few ms; 200 ms collapses them and waits for the
// file to settle before validating.
const debounceWindow = 200 * time.Millisecond

// Watcher watches a single board.yaml for external changes and fans out
// reload notifications to subscribers. The watcher validates each change by
// parsing the YAML; invalid intermediate states do not produce
// notifications and leave the previously-loaded board untouched.
//
// The watch is installed on the file's *parent directory* and filtered to
// the target basename so atomic-rename writes (rename of a temp file over
// the target) are detected — those produce no events when the file itself
// is the watched inode.
type Watcher struct {
	path string
	fsw  *fsnotify.Watcher

	mu   sync.Mutex
	subs []chan struct{}
}

// NewWatcher constructs a Watcher for the given path. The file's parent
// directory must exist; the file itself need not.
func NewWatcher(path string) (*Watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	if err := fsw.Add(dir); err != nil {
		fsw.Close()
		return nil, err
	}
	return &Watcher{path: path, fsw: fsw}, nil
}

// Subscribe returns a buffered channel that receives one struct{} per
// confirmed change, plus a cancel func the caller must invoke (typically
// via defer) to release the subscription. If the consumer is slower than
// the event rate, additional events coalesce — reload is idempotent, so a
// single late wake-up is enough to catch up.
func (w *Watcher) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	w.mu.Lock()
	w.subs = append(w.subs, ch)
	w.mu.Unlock()
	cancel := func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		for i, c := range w.subs {
			if c == ch {
				w.subs = append(w.subs[:i], w.subs[i+1:]...)
				close(c)
				return
			}
		}
	}
	return ch, cancel
}

// Run blocks until ctx is canceled, dispatching debounced fsnotify events
// for the target file to subscribers. The underlying fsnotify.Watcher is
// closed on return.
func (w *Watcher) Run(ctx context.Context) error {
	defer w.fsw.Close()

	base := filepath.Base(w.path)
	var (
		timer  *time.Timer
		timerC <-chan time.Time
	)
	startTimer := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.NewTimer(debounceWindow)
		timerC = timer.C
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case ev, ok := <-w.fsw.Events:
			if !ok {
				return nil
			}
			if filepath.Base(ev.Name) != base {
				continue
			}
			startTimer()

		case err, ok := <-w.fsw.Errors:
			if !ok {
				return nil
			}
			log.Printf("lkan watch: %v", err)

		case <-timerC:
			timerC = nil
			w.validateAndBroadcast()
		}
	}
}

// validateAndBroadcast re-reads and parses the target file. On parse
// success it notifies all subscribers. On any failure (missing file,
// invalid YAML) it logs and returns without broadcasting; the previous
// in-memory board state is preserved by the Store.
func (w *Watcher) validateAndBroadcast() {
	data, err := os.ReadFile(w.path)
	if err != nil {
		log.Printf("lkan watch: read %s: %v", w.path, err)
		return
	}
	var b Board
	if err := yaml.Unmarshal(data, &b); err != nil {
		log.Printf("lkan watch: parse %s: %v (keeping previous board)", w.path, err)
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
