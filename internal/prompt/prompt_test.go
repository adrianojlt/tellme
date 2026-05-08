package prompt

import (
	"strings"
	"testing"

	"tellme/internal/domain"
)

func TestParseResponseValid(t *testing.T) {
	
	raw := `{"suggestions":[{"title":"List files","command":"ls -la","description":"List all files","risk_level":"low","requires_sudo":false,"platform":"any"}]}`
	suggestions, err := ParseResponse(raw)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	if suggestions[0].Command != "ls -la" {
		t.Errorf("expected command 'ls -la', got %q", suggestions[0].Command)
	}
}

func TestParseResponseInvalidJSON(t *testing.T) {

	_, err := ParseResponse("not json at all")

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "invalid JSON response") {
		t.Errorf("error should mention 'invalid JSON response', got: %v", err)
	}
}

func TestParseResponseEmptySuggestions(t *testing.T) {

	_, err := ParseResponse(`{"suggestions":[]}`)

	if err == nil {
		t.Fatal("expected error for empty suggestions, got nil")
	}

	if !strings.Contains(err.Error(), "no suggestions") {
		t.Errorf("error should mention 'no suggestions', got: %v", err)
	}
}

func TestUserPromptContainsQuery(t *testing.T) {

	req := domain.SuggestRequest{UserQuery: "list docker containers", MaxOptions: 3}

	p := UserPrompt(req)

	if !strings.Contains(p, "list docker containers") {
		t.Errorf("prompt missing query: %q", p)
	}

	if !strings.Contains(p, "3") {
		t.Errorf("prompt missing max options: %q", p)
	}
}

func TestUserPromptWithHints(t *testing.T) {

	req := domain.SuggestRequest{
		UserQuery: "test",
		OSHint:    "macos",
		ShellHint: "zsh",
	}

	p := UserPrompt(req)

	if !strings.Contains(p, "macos") {
		t.Errorf("prompt missing OS hint: %q", p)
	}

	if !strings.Contains(p, "zsh") {
		t.Errorf("prompt missing shell hint: %q", p)
	}
}
