package app

import (
	"context"
	"fmt"

	"tellme/internal/domain"
	"tellme/internal/llm"
)

type App struct {
	provider   llm.Provider
	display    func([]domain.CommandSuggestion)
	selectFn   func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error)
	maxOptions int
}

func New(p llm.Provider, display func([]domain.CommandSuggestion), selectFn func([]domain.CommandSuggestion) (*domain.CommandSuggestion, error), maxOptions int) *App {
	return &App{provider: p, display: display, selectFn: selectFn, maxOptions: maxOptions}
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

	return nil
}
