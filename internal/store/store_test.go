package store

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"tellme/internal/domain"
)

func TestLoadMissingFile(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nonexistent-store.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(s.History) != 0 {
		t.Fatalf("expected empty history, got: %+v", s.History)
	}
	if s.Lists == nil {
		t.Fatal("expected non-nil Lists map")
	}
	if len(s.Lists) != 0 {
		t.Fatalf("expected empty Lists map, got: %+v", s.Lists)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "store.json")
	ts := time.Date(2026, 6, 21, 10, 30, 0, 0, time.UTC)

	want := &Store{
		History: []Entry{
			{Query: "list files", Command: "ls -la", Title: "List files", Timestamp: ts},
			{Query: "current dir", Command: "pwd", Title: "Print working dir", Timestamp: ts.Add(time.Hour)},
		},
		Lists: map[string][]Entry{
			"favorites": {
				{Query: "disk usage", Command: "df -h", Title: "Disk usage", Timestamp: ts},
			},
		},
	}

	if err := Save(path, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestSaveUsesTempFileNoLeftover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.json")

	if err := Save(path, empty()); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "store.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("expected only store.json, got: %v", names)
	}
}

func TestAddHistoryBumpsDuplicateCommandToTop(t *testing.T) {
	s := &Store{
		History: []Entry{
			{Command: "ls", Title: "List"},
			{Command: "pwd", Title: "Print dir"},
		},
		Lists: map[string][]Entry{},
	}

	s.AddHistory(Entry{Command: "ls", Title: "List again"})

	if len(s.History) != 2 {
		t.Fatalf("expected history length 2, got %d", len(s.History))
	}
	if s.History[0].Command != "ls" {
		t.Fatalf("expected 'ls' at index 0, got %q", s.History[0].Command)
	}
	if s.History[0].Title != "List again" {
		t.Fatalf("expected bumped entry to carry new title, got %q", s.History[0].Title)
	}
}

func TestAddHistoryCapsAtMax(t *testing.T) {
	s := empty()
	for i := 0; i < maxHistory; i++ {
		s.AddHistory(Entry{Command: "cmd-" + strconv.Itoa(i)})
	}
	oldest := s.History[len(s.History)-1].Command

	s.AddHistory(Entry{Command: "fresh"})

	if len(s.History) != maxHistory {
		t.Fatalf("expected history length %d, got %d", maxHistory, len(s.History))
	}
	if s.History[0].Command != "fresh" {
		t.Fatalf("expected 'fresh' at index 0, got %q", s.History[0].Command)
	}
	for _, h := range s.History {
		if h.Command == oldest {
			t.Fatalf("expected oldest entry %q to be dropped", oldest)
		}
	}
}

func TestAddToListRejectsDuplicateCommand(t *testing.T) {
	s := &Store{
		Lists: map[string][]Entry{
			"work": {{Command: "git status", Title: "Status"}},
		},
	}

	err := s.AddToList("work", Entry{Command: "git status", Title: "Other"})

	if err == nil {
		t.Fatal("expected non-nil error for duplicate command, got nil")
	}
	if len(s.Lists["work"]) != 1 {
		t.Fatalf("expected list unchanged at length 1, got %d", len(s.Lists["work"]))
	}
	if s.Lists["work"][0].Title != "Status" {
		t.Fatalf("expected original entry preserved, got %q", s.Lists["work"][0].Title)
	}
}

func TestAddToListCreatesListWhenAbsent(t *testing.T) {
	s := empty()

	if err := s.AddToList("new", Entry{Command: "ls"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Lists["new"]) != 1 {
		t.Fatalf("expected list 'new' with 1 entry, got %d", len(s.Lists["new"]))
	}
}

func TestRecentHistoryTruncatesToN(t *testing.T) {
	s := &Store{
		History: []Entry{
			{Command: "a"},
			{Command: "b"},
			{Command: "c"},
		},
	}

	got := s.RecentHistory(2)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Command != "a" || got[1].Command != "b" {
		t.Fatalf("expected newest two [a b], got %+v", got)
	}

	all := s.RecentHistory(10)
	if len(all) != 3 {
		t.Fatalf("expected all 3 entries when n exceeds length, got %d", len(all))
	}
}

func TestSuggestionMappingPreservesTitleAndCommand(t *testing.T) {
	e := Entry{
		Query:     "list files",
		Command:   "ls -la",
		Title:     "List files",
		Timestamp: time.Date(2026, 6, 21, 10, 30, 0, 0, time.UTC),
	}

	sug := e.ToSuggestion()
	if sug.Title != e.Title || sug.Command != e.Command {
		t.Fatalf("ToSuggestion lost data: got %+v", sug)
	}
	if sug.Description != "" || sug.RiskLevel != "" || sug.RequiresSudo || sug.Platform != "" {
		t.Fatalf("expected only Title+Command populated, got %+v", sug)
	}

	back := FromSuggestion(sug)
	if back.Title != e.Title || back.Command != e.Command {
		t.Fatalf("round-trip lost Title/Command: got %+v", back)
	}
	if !back.Timestamp.IsZero() || back.Query != "" {
		t.Fatalf("expected Query/Timestamp empty on the way in, got %+v", back)
	}

	if FromSuggestion(domain.CommandSuggestion{Command: "x"}).Command != "x" {
		t.Fatal("FromSuggestion failed on bare suggestion")
	}
}

func TestLoadCorruptJSONLeavesFileUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	corrupt := []byte(`{"history": [ not valid json`)
	if err := os.WriteFile(path, corrupt, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}

	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !reflect.DeepEqual(after, corrupt) {
		t.Fatalf("file content changed: got %q, want %q", after, corrupt)
	}
}
