package lkan

import (
	"net/http"
)

// Option configures NewServer. Use WithWatcher to opt in to live-reload
// behavior; the zero-option call yields the same server as before.
type Option func(*serverOpts)

type serverOpts struct {
	watcher *Watcher
}

// WithWatcher attaches a Watcher to the server, enabling the GET /events SSE
// stream and the live-reload data attribute on the rendered page.
func WithWatcher(w *Watcher) Option { return func(o *serverOpts) { o.watcher = w } }

// NewServer returns an http.Handler with all lkan routes wired up against
// the given store.
func NewServer(s *Store, opts ...Option) http.Handler {
	o := serverOpts{}
	for _, opt := range opts {
		opt(&o)
	}
	mux := http.NewServeMux()
	h := &handlers{store: s, watcher: o.watcher}

	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /api/board", h.getBoard)
	mux.HandleFunc("POST /api/cards", h.createCard)
	mux.HandleFunc("PATCH /api/cards/{id}", h.editCard)
	mux.HandleFunc("DELETE /api/cards/{id}", h.deleteCard)
	mux.HandleFunc("POST /api/cards/{id}/move", h.moveCard)
	if o.watcher != nil {
		mux.HandleFunc("GET /events", h.events)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(StaticFS()))))

	return mux
}
