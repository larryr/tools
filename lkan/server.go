package lkan

import (
	"net/http"
)

// NewServer returns an http.Handler with all lkan routes wired up against
// the given store.
func NewServer(s *Store) http.Handler {
	mux := http.NewServeMux()
	h := &handlers{store: s}

	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /api/board", h.getBoard)
	mux.HandleFunc("POST /api/cards", h.createCard)
	mux.HandleFunc("PATCH /api/cards/{id}", h.editCard)
	mux.HandleFunc("DELETE /api/cards/{id}", h.deleteCard)
	mux.HandleFunc("POST /api/cards/{id}/move", h.moveCard)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(StaticFS()))))

	return mux
}
