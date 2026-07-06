package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"tellme/internal/store"
)

func TestIsListNames(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"bare list", []string{"list"}, true},
		{"list with name", []string{"list", "work"}, false},
		{"list multi-word query", []string{"list", "active", "containers"}, false},
		{"no args", []string{}, false},
		{"not list keyword", []string{"history"}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsListNames(c.args); got != c.want {
				t.Errorf("IsListNames(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

func TestIsListNamed(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"list with name", []string{"list", "work"}, true},
		{"bare list", []string{"list"}, false},
		{"list multi-word query", []string{"list", "active", "containers"}, false},
		{"two args not list", []string{"history", "work"}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := IsListNamed(c.args); got != c.want {
				t.Errorf("IsListNamed(%v) = %v, want %v", c.args, got, c.want)
			}
		})
	}
}

// captureStdout runs fn while capturing everything written to os.Stdout and
// returns it. Used to assert on the printed names/counts and messages.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	w.Close()
	os.Stdout = orig
	return <-done
}

func TestRunListNamesPrintsNamesAndCounts(t *testing.T) {
	lists := map[string][]store.Entry{
		"favorites": {{Command: "ls"}, {Command: "pwd"}},
		"work":      {{Command: "docker ps"}},
	}

	out := captureStdout(t, func() {
		if err := printListNames(lists); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "favorites (2)") {
		t.Errorf("expected %q to contain favorites count, got %q", out, "favorites (2)")
	}
	if !strings.Contains(out, "work (1)") {
		t.Errorf("expected %q to contain work count, got %q", out, "work (1)")
	}
}

func TestRunListNamesEmptyPrintsFriendlyMessage(t *testing.T) {
	out := captureStdout(t, func() {
		if err := printListNames(map[string][]store.Entry{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No lists yet.") {
		t.Errorf("expected friendly message, got %q", out)
	}
}

func TestRunListNamedSelectsAndRuns(t *testing.T) {
	entries := []store.Entry{
		{Title: "list containers", Command: "docker ps"},
		{Title: "images", Command: "docker images"},
	}

	actions, ran, _, _, _ := noopActions()

	// Select index 1 (docker ps), then action 1 (Run).
	if err := runListing(strings.NewReader("1\n1\n"), entries, "work", actions); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *ran != "docker ps" {
		t.Errorf("expected Run of %q, got %q", "docker ps", *ran)
	}
}

func TestRunListUnknownPrintsNoSuchListAndExitsZero(t *testing.T) {
	lists := map[string][]store.Entry{"work": {{Command: "docker ps"}}}

	var out string
	var err error
	out = captureStdout(t, func() {
		err = lookupList(lists, "nope", func() error { return nil })
	})

	if err != nil {
		t.Fatalf("expected nil error for unknown list, got %v", err)
	}
	if !strings.Contains(out, "No such list: nope") {
		t.Errorf("expected no-such-list message, got %q", out)
	}
}
