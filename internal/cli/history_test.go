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

// noopActions returns a ListActions whose callbacks all record their command
// into the corresponding string pointer (nil for unused actions). Tests can
// pass a single set of pointers and read what was invoked.
func noopActions() (ListActions, *string, *string, *string, *string) {

	var ran, ranAndCopied, copied, removed string

	actions := ListActions{
		Run:        func(c string) error { ran = c; return nil },
		RunAndCopy: func(c string) error { ranAndCopied = c; return nil },
		Copy:       func(c string) error { copied = c; return nil },
		Remove:     func(c string) error { removed = c; return nil },
	}

	return actions, &ran, &ranAndCopied, &copied, &removed
}

func TestRunListingSelectsAndRuns(t *testing.T) {
	entries := []store.Entry{
		{Title: "list files", Command: "ls -la"},
		{Title: "disk usage", Command: "df -h"},
	}

	actions, ran, _, _, _ := noopActions()

	// Select index 2 (df -h), then action 1 (Run).
	err := runListing(strings.NewReader("2\n1\n"), entries, "history", actions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *ran != "df -h" {
		t.Errorf("expected Run of %q, got %q", "df -h", *ran)
	}
}

func TestRunListingSelectsRunsAndCopies(t *testing.T) {
	entries := []store.Entry{{Title: "list files", Command: "ls -la"}}

	actions, _, ranAndCopied, _, _ := noopActions()

	// Select item 1, then action 2 (Run and Copy).
	if err := runListing(strings.NewReader("1\n2\n"), entries, "favorites", actions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *ranAndCopied != "ls -la" {
		t.Errorf("expected RunAndCopy of %q, got %q", "ls -la", *ranAndCopied)
	}
}

func TestRunListingSelectsAndCopies(t *testing.T) {
	entries := []store.Entry{{Title: "list files", Command: "ls -la"}}

	actions, _, _, copied, _ := noopActions()

	// Select item 1, then action 3 (Copy).
	if err := runListing(strings.NewReader("1\n3\n"), entries, "favorites", actions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *copied != "ls -la" {
		t.Errorf("expected Copy of %q, got %q", "ls -la", *copied)
	}
}

func TestRunListingSelectsAndRemoves(t *testing.T) {
	entries := []store.Entry{{Title: "list files", Command: "ls -la"}}

	actions, _, _, _, removed := noopActions()

	// Select item 1, then action 4 (Remove).
	if err := runListing(strings.NewReader("1\n4\n"), entries, "favorites", actions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *removed != "ls -la" {
		t.Errorf("expected Remove of %q, got %q", "ls -la", *removed)
	}
}

func TestRunListingExitSelectionDoesNotInvokeAnyAction(t *testing.T) {
	entries := []store.Entry{{Title: "list files", Command: "ls -la"}}

	actions, ran, ranAndCopied, copied, removed := noopActions()

	if err := runListing(strings.NewReader("0\n"), entries, "history", actions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *ran != "" || *ranAndCopied != "" || *copied != "" || *removed != "" {
		t.Errorf("expected no action invoked, got run=%q runAndCopy=%q copy=%q remove=%q",
			*ran, *ranAndCopied, *copied, *removed)
	}
}

func TestRunListingEmptyPrintsFriendlyMessageAndReturnsNil(t *testing.T) {
	actions, ran, _, _, _ := noopActions()

	if err := runListing(strings.NewReader(""), nil, "favorites", actions); err != nil {
		t.Fatalf("expected nil error for empty listing, got %v", err)
	}

	if *ran != "" {
		t.Errorf("expected Run not to be called for empty listing, got %q", *ran)
	}
}
