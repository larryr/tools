package lkan

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (*httptest.Server, *Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "board.yaml")
	writeBoard(t, path, sampleBoard())
	store, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(NewServer(store))
	t.Cleanup(srv.Close)
	return srv, store, path
}

func TestHandler_Index(t *testing.T) {
	srv, _, _ := newTestServer(t)
	res, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type %q", ct)
	}
}

func TestHandler_StaticServed(t *testing.T) {
	srv, _, _ := newTestServer(t)
	res, err := http.Get(srv.URL + "/static/kanban.css")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestHandler_MoveCard(t *testing.T) {
	srv, store, _ := newTestServer(t)
	body, _ := json.Marshal(moveReq{Column: "doing", Index: 0})
	res, err := http.Post(srv.URL+"/api/cards/b/move", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	snap, _ := store.Snapshot()
	if got := cardIDs(snap.Columns[1]); !eq(got, []string{"b", "d"}) {
		t.Errorf("doing: %v", got)
	}
}

func TestHandler_MoveCard_NotFound(t *testing.T) {
	srv, _, _ := newTestServer(t)
	body, _ := json.Marshal(moveReq{Column: "doing", Index: 0})
	res, err := http.Post(srv.URL+"/api/cards/nope/move", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestHandler_MoveCard_BadJSON(t *testing.T) {
	srv, _, _ := newTestServer(t)
	res, err := http.Post(srv.URL+"/api/cards/a/move", "application/json", strings.NewReader("not json"))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestHandler_CreateCard(t *testing.T) {
	srv, store, _ := newTestServer(t)
	body, _ := json.Marshal(createCardReq{Column: "todo", Title: "fresh"})
	res, err := http.Post(srv.URL+"/api/cards", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got map[string]string
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["id"] == "" {
		t.Error("expected id in response")
	}
	snap, _ := store.Snapshot()
	if n := len(snap.Columns[0].Cards); n != 4 {
		t.Errorf("expected 4 cards in todo, got %d", n)
	}
}

func TestHandler_EditCard(t *testing.T) {
	srv, store, _ := newTestServer(t)
	title := "renamed"
	body, _ := json.Marshal(editCardReq{Title: &title})
	req, _ := http.NewRequest("PATCH", srv.URL+"/api/cards/a", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	snap, _ := store.Snapshot()
	c, _, _ := snap.FindCard("a")
	if c.Title != "renamed" {
		t.Errorf("title: %q", c.Title)
	}
}

func TestHandler_DeleteCard(t *testing.T) {
	srv, store, _ := newTestServer(t)
	req, _ := http.NewRequest("DELETE", srv.URL+"/api/cards/a", nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	snap, _ := store.Snapshot()
	if c, _, _ := snap.FindCard("a"); c != nil {
		t.Error("card a still present")
	}
}

func TestHandler_GetBoardJSON(t *testing.T) {
	srv, _, _ := newTestServer(t)
	res, err := http.Get(srv.URL + "/api/board")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	var b Board
	if err := json.NewDecoder(res.Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	if len(b.Columns) != 3 {
		t.Errorf("columns: %d", len(b.Columns))
	}
}
