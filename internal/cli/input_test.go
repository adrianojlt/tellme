package cli

import (
	"strings"
	"testing"

	"tellme/internal/domain"
)

var testSuggestions = []domain.CommandSuggestion{
	{Title: "A", Command: "cmd-a"},
	{Title: "B", Command: "cmd-b"},
}

func TestSelectFromValidChoice(t *testing.T) {

	got, err := selectFrom(strings.NewReader("2\n"), testSuggestions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected non-nil suggestion")
	}

	if got.Command != "cmd-b" {
		t.Errorf("expected cmd-b, got %q", got.Command)
	}
}

func TestSelectFromZeroReturnsNil(t *testing.T) {

	got, err := selectFrom(strings.NewReader("0\n"), testSuggestions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil for choice 0, got %+v", got)
	}
}

func TestSelectFromEOFReturnsNil(t *testing.T) {

	got, err := selectFrom(strings.NewReader(""), testSuggestions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != nil {
		t.Errorf("expected nil on EOF, got %+v", got)
	}
}

func TestSelectFromInvalidThenValid(t *testing.T) {

	got, err := selectFrom(strings.NewReader("abc\n1\n"), testSuggestions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || got.Command != "cmd-a" {
		t.Errorf("expected cmd-a, got %+v", got)
	}
}

func TestSelectFromOutOfRangeThenValid(t *testing.T) {

	got, err := selectFrom(strings.NewReader("5\n1\n"), testSuggestions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || got.Command != "cmd-a" {
		t.Errorf("expected cmd-a, got %+v", got)
	}
}
