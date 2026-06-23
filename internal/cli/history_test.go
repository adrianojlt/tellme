package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"tellme/internal/store"
)

// listContainsCommand reports whether the named list in the reloaded store
// contains an entry with the given command.
func listContainsCommand(t *testing.T, storePath, list, command string) bool {
	t.Helper()
	s, err := store.Load(storePath)
	if err != nil {
		t.Fatalf("reload store: %v", err)
	}
	for _, e := range s.Lists[list] {
		if e.Command == command {
			return true
		}
	}
	return false
}

func noReRun(t *testing.T) func(string) error {
	t.Helper()
	return func(string) error {
		t.Fatalf("reRun should not be called for an add action")
		return nil
	}
}

func TestRunHistoryAddToFavorites(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	s := &store.Store{Lists: map[string][]store.Entry{}}
	entries := []store.Entry{{Title: "list files", Command: "ls -la", Query: "list files"}}

	// Select item 1, then action 2 (Add to Favorites).
	in := strings.NewReader("1\n2\n")
	if err := runHistoryListing(in, s, storePath, entries, noReRun(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !listContainsCommand(t, storePath, "favorites", "ls -la") {
		t.Error("expected favorites list to contain the selected command")
	}
}

func TestRunHistoryCreateNewList(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	s := &store.Store{Lists: map[string][]store.Entry{}}
	entries := []store.Entry{{Title: "disk usage", Command: "df -h"}}

	// Select item 1, action 3 (Add to a List), pick "create new" (option 1 since
	// no lists exist), then type the name "work".
	in := strings.NewReader("1\n3\n1\nwork\n")
	if err := runHistoryListing(in, s, storePath, entries, noReRun(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !listContainsCommand(t, storePath, "work", "df -h") {
		t.Error("expected 'work' list to be created with the selected command")
	}
}

func TestRunHistoryAddDuplicateRejected(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	s := &store.Store{Lists: map[string][]store.Entry{
		"work": {{Title: "disk usage", Command: "df -h"}},
	}}
	entries := []store.Entry{{Title: "disk usage", Command: "df -h"}}

	// Select item 1, action 3, pick existing "work" (option 1).
	in := strings.NewReader("1\n3\n1\n")
	if err := runHistoryListing(in, s, storePath, entries, noReRun(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The duplicate is rejected before saving, so the list is unchanged and must
	// still contain exactly one entry.
	if got := len(s.Lists["work"]); got != 1 {
		t.Errorf("expected 'work' to be unchanged with 1 entry, got %d", got)
	}
}

func TestRunHistoryCreateNewListBlankNameRejected(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "store.json")
	s := &store.Store{Lists: map[string][]store.Entry{}}
	entries := []store.Entry{{Title: "disk usage", Command: "df -h"}}

	// Select item 1, action 3, "create new" (option 1), then a whitespace name.
	in := strings.NewReader("1\n3\n1\n   \n")
	if err := runHistoryListing(in, s, storePath, entries, noReRun(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.Lists) != 0 {
		t.Errorf("expected no lists to be created, got %v", s.Lists)
	}
}

func TestIsSubcommandExactMatch(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"history keyword", []string{"history"}, true},
		{"favorites keyword", []string{"favorites"}, true},
		{"query first token find", []string{"find big files"}, false},
		{"bare find word", []string{"find"}, false},
		{"history with extra arg", []string{"history", "extra"}, false},
		{"no args", []string{}, false},
		{"flag", []string{"--config"}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsSubcommand(c.args); got != c.want {
				t.Errorf("IsSubcommand(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

func TestRunListingSelectsAndReRuns(t *testing.T) {
	entries := []store.Entry{
		{Title: "list files", Command: "ls -la"},
		{Title: "disk usage", Command: "df -h"},
	}

	var ran string
	reRun := func(command string) error {
		ran = command
		return nil
	}
	noRemove := func(string) error { return nil }

	// Select index 2 (df -h), then action 1 (run/copy).
	err := runListing(strings.NewReader("2\n1\n"), entries, "history", reRun, noRemove)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ran != "df -h" {
		t.Errorf("expected re-run of %q, got %q", "df -h", ran)
	}
}

func TestRunListingExitSelectionDoesNotReRun(t *testing.T) {
	entries := []store.Entry{{Title: "list files", Command: "ls -la"}}

	called := false
	reRun := func(command string) error {
		called = true
		return nil
	}
	noRemove := func(string) error { return nil }

	if err := runListing(strings.NewReader("0\n"), entries, "history", reRun, noRemove); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("expected reRun not to be called on exit selection")
	}
}

func TestRunListingEmptyPrintsFriendlyMessageAndReturnsNil(t *testing.T) {
	called := false
	reRun := func(command string) error {
		called = true
		return nil
	}
	noRemove := func(string) error { return nil }

	if err := runListing(strings.NewReader(""), nil, "favorites", reRun, noRemove); err != nil {
		t.Fatalf("expected nil error for empty listing, got %v", err)
	}

	if called {
		t.Error("expected reRun not to be called for empty listing")
	}
}
