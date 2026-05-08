package fake

import (
	"context"

	"tellme/internal/domain"
)

type Provider struct{}

func New() *Provider {
	return &Provider{}
}

func (p *Provider) SuggestCommands(_ context.Context, _ domain.SuggestRequest) ([]domain.CommandSuggestion, error) {
	return []domain.CommandSuggestion{
		{
			Title:        "List running containers",
			Command:      "docker ps",
			Description:  "Shows all currently running Docker containers",
			RiskLevel:    "low",
			RequiresSudo: false,
			Platform:     "any",
		},
		{
			Title:        "List all containers",
			Command:      "docker ps -a",
			Description:  "Shows all containers, including stopped ones",
			RiskLevel:    "low",
			RequiresSudo: false,
			Platform:     "any",
		},
		{
			Title:        "List containers with size",
			Command:      "docker ps -s",
			Description:  "Shows running containers with their disk usage",
			RiskLevel:    "low",
			RequiresSudo: false,
			Platform:     "any",
		},
	}, nil
}
