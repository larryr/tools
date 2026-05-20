package lkan

import "net/http"

// events serves a Server-Sent Events stream that emits a "reload" event
// every time the watcher detects a change to board.yaml. Returns 404 when
// the server was constructed without a Watcher.
func (h *handlers) events(w http.ResponseWriter, r *http.Request) {
	if h.watcher == nil {
		http.NotFound(w, r)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(": ok\n\n")); err != nil {
		return
	}
	flusher.Flush()

	ch, cancel := h.watcher.Subscribe()
	defer cancel()

	for {
		select {
		case <-r.Context().Done():
			return
		case _, ok := <-ch:
			if !ok {
				return
			}
			if _, err := w.Write([]byte("event: reload\ndata: {}\n\n")); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
