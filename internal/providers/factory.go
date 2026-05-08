package providers

import (
	"fmt"

	"tellme/internal/config"
	"tellme/internal/llm"
	"tellme/internal/providers/anthropic"
)

func New(cfg *config.Config) (llm.Provider, error) {

	switch cfg.Provider {

	case "anthropic":

		key := cfg.Providers.Anthropic.APIKey

		if key == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
		}

		return anthropic.New(key, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unknown provider %q; supported: anthropic", cfg.Provider)
	}
}
