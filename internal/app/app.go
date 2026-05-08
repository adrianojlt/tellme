package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"tellme/internal/domain"
	"tellme/internal/llm"
)

type App struct {
	provider        llm.Provider
	display         func([]domain.CommandSuggestion)
	selectFn        func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error)
	copyFn          func(string) error
	maxOptions      int
	copyAfterSelect bool
}

func New(
	p llm.Provider,
	display func([]domain.CommandSuggestion),
	selectFn func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error),
	maxOptions int,
	copyFn func(string) error,
	copyAfterSelect bool,
) *App {
	return &App{
		provider:        p,
		display:         display,
		selectFn:        selectFn,
		maxOptions:      maxOptions,
		copyFn:          copyFn,
		copyAfterSelect: copyAfterSelect,
	}
}

func (a *App) Run(ctx context.Context, query string) error {
	suggestions, err := a.provider.SuggestCommands(ctx, domain.SuggestRequest{
		UserQuery:  query,
		MaxOptions: a.maxOptions,
	})
	if err != nil {
		return err
	}

	a.display(suggestions)

	selected, err := a.selectFn(suggestions)
	if err != nil {
		return err
	}

	if selected == nil {
		fmt.Println("Bye.")
		return nil
	}

	fmt.Printf("Selected: %s\n", selected.Command)
	a.handleClipboard(selected.Command)

	return nil
}

func (a *App) handleClipboard(command string) {

	if a.copyAfterSelect {

		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
		} else {
			fmt.Println("Copied to clipboard.")
		}

		return
	}

	fmt.Print("Copy to clipboard? (y/n): ")

	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')

	if strings.TrimSpace(line) == "y" {

		if err := a.copyFn(command); err != nil {
			fmt.Printf("Clipboard error: %v\n", err)
		} else {
			fmt.Println("Copied to clipboard.")
		}
	}
}
