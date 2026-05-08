package main

import (
	"context"
	"fmt"
	"os"

	"tellme/internal/domain"
	"tellme/internal/providers/fake"
)

const usage = `Usage: tellme <query>

Ask a natural language question and get shell command suggestions.

Example:
  tellme "list active docker containers"
`

func main() {
	args := os.Args[1:]

	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Print(usage)
		os.Exit(0)
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	query := args[0]

	provider := fake.New()
	suggestions, err := provider.SuggestCommands(context.Background(), domain.SuggestRequest{
		UserQuery: query,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, s := range suggestions {
		fmt.Printf("%s: %s\n", s.Title, s.Command)
	}
}
