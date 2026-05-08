package llm

import (
	"context"

	"tellme/internal/domain"
)

type Provider interface {
	SuggestCommands(ctx context.Context, req domain.SuggestRequest) ([]domain.CommandSuggestion, error)
}
