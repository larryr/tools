package lkan

import (
	"encoding/json"
	"errors"
	"net/http"
)

type handlers struct {
	store   *Store
	watcher *Watcher // nil when --watch is not set; turns on /events and liveReload
}

// pageData wraps a Board with a flag the template uses to opt the page in to
// the live-reload SSE stream. Board is embedded so existing template
// expressions like {{.Title}} continue to resolve.
type pageData struct {
	*Board
	LiveReload bool
}

func (h *handlers) index(w http.ResponseWriter, r *http.Request) {
	b, err := h.store.Snapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := boardTmpl.Execute(w, pageData{Board: b, LiveReload: h.watcher != nil}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *handlers) getBoard(w http.ResponseWriter, r *http.Request) {
	b, err := h.store.Snapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

type createCardReq struct {
	Column      string   `json:"column"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Assignees   []string `json:"assignees,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Due         string   `json:"due,omitempty"`
}

func (h *handlers) createCard(w http.ResponseWriter, r *http.Request) {
	var req createCardReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	var newID string
	err := h.store.Update(func(b *Board) error {
		id, err := AddCard(b, req.Column, Card{
			Title:       req.Title,
			Description: req.Description,
			Assignees:   req.Assignees,
			Labels:      req.Labels,
			Due:         req.Due,
		})
		newID = id
		return err
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": newID})
}

type editCardReq struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Assignees   *[]string `json:"assignees,omitempty"`
	Labels      *[]string `json:"labels,omitempty"`
	Due         *string   `json:"due,omitempty"`
}

func (h *handlers) editCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req editCardReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	err := h.store.Update(func(b *Board) error {
		return EditCard(b, id, CardPatch{
			Title:       req.Title,
			Description: req.Description,
			Assignees:   req.Assignees,
			Labels:      req.Labels,
			Due:         req.Due,
		})
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) deleteCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.store.Update(func(b *Board) error {
		return DeleteCard(b, id)
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type moveReq struct {
	Column string `json:"column"`
	Index  int    `json:"index"`
}

func (h *handlers) moveCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req moveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	err := h.store.Update(func(b *Board) error {
		return MoveCard(b, id, req.Column, req.Index)
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
