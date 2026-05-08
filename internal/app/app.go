package app

import (
	"context"

	"tellme/internal/domain"
	"tellme/internal/llm"
)

type App struct {
	provider llm.Provider
	display  func([]domain.CommandSuggestion)
}

func New(p llm.Provider, display func([]domain.CommandSuggestion)) *App {
	return &App{provider: p, display: display}
}

func (a *App) Run(ctx context.Context, query string) error {
	suggestions, err := a.provider.SuggestCommands(ctx, domain.SuggestRequest{
		UserQuery: query,
	})
	if err != nil {
		return err
	}

	a.display(suggestions)
	return nil
}
