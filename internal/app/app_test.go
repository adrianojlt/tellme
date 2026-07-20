package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tellme/internal/store"
)

var errClipboard = errors.New("clipboard unavailable")

func newTestApp(copyFn func(string) error, copyAfterSelect bool) *App {
	return &App{
		copyFn:          copyFn,
		copyAfterSelect: copyAfterSelect,
	}
}

func TestRunSelectedCopyBranch(t *testing.T) {

	var copied []string
	a := newTestApp(func(c string) error {
		copied = append(copied, c)
		return nil
	}, false)

	err := a.runSelected(strings.NewReader("n\ny\n"), "", "", "echo hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copied) != 1 {
		t.Fatalf("expected copyFn called once, got %d", len(copied))
	}

	if copied[0] != "echo hi" {
		t.Errorf("expected copied %q, got %q", "echo hi", copied[0])
	}
}

func TestRunSelectedExecuteBranch(t *testing.T) {

	var copied []string
	a := newTestApp(func(c string) error {
		copied = append(copied, c)
		return nil
	}, false)

	// "y" runs via shell exec path; use a harmless command.
	err := a.runSelected(strings.NewReader("y\n"), "", "", "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copied) != 0 {
		t.Errorf("expected copyFn not called, got %d calls", len(copied))
	}
}

func TestRunSelectedRecordsOnExecute(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	a := newTestApp(func(string) error { return nil }, false)
	a.storePath = storePath

	err := a.runSelected(strings.NewReader("y\n"), "list files", "List files", "true")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(s.History))
	}
	e := s.History[0]
	if e.Query != "list files" || e.Command != "true" || e.Title != "List files" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if e.Timestamp.IsZero() {
		t.Errorf("expected non-zero timestamp")
	}
	if e.Timestamp.Location().String() != "UTC" {
		t.Errorf("expected UTC timestamp, got %s", e.Timestamp.Location())
	}
}

func TestRunSelectedRecordsOnCopy(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	a := newTestApp(func(string) error { return nil }, false)
	a.storePath = storePath

	err := a.runSelected(strings.NewReader("n\ny\n"), "list files", "List files", "echo hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(s.History))
	}
	e := s.History[0]
	if e.Query != "list files" || e.Command != "echo hi" || e.Title != "List files" {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestRunSelectedDedupOnRepeat(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	a := newTestApp(func(string) error { return nil }, false)
	a.storePath = storePath

	if err := a.runSelected(strings.NewReader("y\n"), "q1", "T1", "true"); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := a.runSelected(strings.NewReader("y\n"), "q2", "T2", "true"); err != nil {
		t.Fatalf("second run: %v", err)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Fatalf("expected 1 history entry after dedup, got %d", len(s.History))
	}
	if s.History[0].Query != "q2" {
		t.Errorf("expected bumped entry query %q, got %q", "q2", s.History[0].Query)
	}
}

func TestRunSelectedRecordNonFatalOnWriteError(t *testing.T) {
	// A store path under a non-existent, unwritable parent forces a save error.
	// store.Save does MkdirAll(filepath.Dir(path)); point dir at a file so the
	// mkdir fails.
	dir := t.TempDir()
	notADir := filepath.Join(dir, "blocker")
	if err := os.WriteFile(notADir, []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	storePath := filepath.Join(notADir, "store.json")

	a := newTestApp(func(string) error { return nil }, false)
	a.storePath = storePath

	// Command must still run; runSelected returns nil regardless of store error.
	err := a.runSelected(strings.NewReader("y\n"), "q", "T", "true")
	if err != nil {
		t.Fatalf("expected non-fatal store error, got %v", err)
	}

	// Nothing should have been persisted at the unwritable path.
	if _, statErr := os.Stat(storePath); statErr == nil {
		t.Errorf("expected store file not written, but it exists")
	}
}

func TestRunSelectedNoRecordWhenNotExecuted(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	a := newTestApp(func(string) error { return nil }, false)
	a.storePath = storePath

	// Decline execute, decline copy.
	err := a.runSelected(strings.NewReader("n\nn\n"), "q", "T", "echo hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 0 {
		t.Errorf("expected no history entries, got %d", len(s.History))
	}
}

func TestRunCommandExecutesAndRecords(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	var copied []string
	a := newTestApp(func(c string) error {
		copied = append(copied, c)
		return nil
	}, false)
	a.storePath = storePath

	if err := a.RunCommand("true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copied) != 0 {
		t.Errorf("expected copyFn not called, got %d calls", len(copied))
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(s.History))
	}
	if s.History[0].Command != "true" {
		t.Errorf("expected history command %q, got %q", "true", s.History[0].Command)
	}
}

func TestRunAndCopyCommandExecutesRecordsAndCopies(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	var copied []string
	a := newTestApp(func(c string) error {
		copied = append(copied, c)
		return nil
	}, false)
	a.storePath = storePath

	if err := a.RunAndCopyCommand("true"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copied) != 1 || copied[0] != "true" {
		t.Errorf("expected copyFn called once with %q, got %v", "true", copied)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(s.History))
	}
}

func TestCopyCommandCopiesAndRecords(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	var copied []string
	a := newTestApp(func(c string) error {
		copied = append(copied, c)
		return nil
	}, false)
	a.storePath = storePath

	if err := a.CopyCommand("echo hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(copied) != 1 || copied[0] != "echo hi" {
		t.Errorf("expected copyFn called once with %q, got %v", "echo hi", copied)
	}

	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("load store: %v", err)
	}
	if len(s.History) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(s.History))
	}
	if s.History[0].Command != "echo hi" {
		t.Errorf("expected history command %q, got %q", "echo hi", s.History[0].Command)
	}
}

func TestCopyCommandDoesNotRecordOnClipboardFailure(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	a := newTestApp(func(string) error { return errClipboard }, false)
	a.storePath = storePath

	err := a.CopyCommand("echo hi")
	if err == nil {
		t.Fatal("expected error from CopyCommand, got nil")
	}

	if _, statErr := os.Stat(storePath); statErr == nil {
		t.Errorf("expected store file not created on clipboard failure, but it exists")
	}
}

func TestCopyCommandReturnsErrorOnClipboardFailure(t *testing.T) {
	a := newTestApp(func(string) error { return errClipboard }, false)

	err := a.CopyCommand("echo hi")
	if err == nil {
		t.Fatal("expected error from CopyCommand, got nil")
	}
	if !strings.Contains(err.Error(), "clipboard") {
		t.Errorf("expected error to mention clipboard, got %v", err)
	}
}
