package providers

import (
	"fmt"

	"tellme/internal/config"
	"tellme/internal/llm"
	"tellme/internal/providers/anthropic"
	"tellme/internal/providers/groq"
	"tellme/internal/providers/mistral"
	"tellme/internal/providers/openai"
)

func New(cfg *config.Config) (llm.Provider, error) {
	switch cfg.Provider {

	case "anthropic":

		key := cfg.Providers.Anthropic.APIKey

		if key == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
		}

		return anthropic.New(key, cfg.Model), nil

	case "openai":

		key := cfg.Providers.OpenAI.APIKey

		if key == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is not set")
		}

		return openai.New(key, cfg.Model), nil

	case "mistral":

		key := cfg.Providers.Mistral.APIKey

		if key == "" {
			return nil, fmt.Errorf("MISTRAL_API_KEY is not set")
		}

		return mistral.New(key, cfg.Model), nil

	case "groq":

		key := cfg.Providers.Groq.APIKey

		if key == "" {
			return nil, fmt.Errorf("GROQ_API_KEY is not set")
		}

		return groq.New(key, cfg.Model), nil

	default:
		return nil, fmt.Errorf("unknown provider %q; supported: anthropic, openai, mistral, groq", cfg.Provider)
	}
}
