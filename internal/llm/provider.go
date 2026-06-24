package llm

import (
	"context"

	"tellme/internal/domain"
)

// Provider asks an LLM for shell command suggestions. Implementations must
// respect context cancellation and deadline; a cancelled or timed-out context
// should surface as a context.Canceled or context.DeadlineExceeded error.
type Provider interface {
	SuggestCommands(ctx context.Context, req domain.SuggestRequest) ([]domain.CommandSuggestion, error)
}
