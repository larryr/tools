package lkan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateAgentsMD_Create(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")

	action, err := UpdateAgentsMD(path)
	if err != nil {
		t.Fatal(err)
	}
	if action != "created" {
		t.Errorf("action = %q, want %q", action, "created")
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), AgentsBeginMarker) || !strings.Contains(string(got), AgentsEndMarker) {
		t.Error("missing markers")
	}
}

func TestUpdateAgentsMD_Append(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	original := "# Project Agents\n\nHuman content.\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	action, err := UpdateAgentsMD(path)
	if err != nil {
		t.Fatal(err)
	}
	if action != "appended" {
		t.Errorf("action = %q, want %q", action, "appended")
	}
	got, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(got), original) {
		t.Error("original content not preserved at top")
	}
	if !strings.Contains(string(got), AgentsBeginMarker) {
		t.Error("missing begin marker")
	}
}

func TestUpdateAgentsMD_ReplaceInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	original := "TOP STAYS\n\n" + AgentsBeginMarker + "\nstale\n" + AgentsEndMarker + "\n\nBOTTOM STAYS\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	action, err := UpdateAgentsMD(path)
	if err != nil {
		t.Fatal(err)
	}
	if action != "updated" {
		t.Errorf("action = %q, want %q", action, "updated")
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "TOP STAYS") || !strings.Contains(string(got), "BOTTOM STAYS") {
		t.Error("surrounding text not preserved")
	}
	if strings.Contains(string(got), "stale") {
		t.Error("stale content not replaced")
	}
}

func TestUpdateAgentsMD_HalfOpenMarkerRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(path, []byte("\n"+AgentsBeginMarker+"\nno end\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := UpdateAgentsMD(path); err == nil {
		t.Fatal("expected error for half-open marker block, got nil")
	}
}
