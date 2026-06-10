package providers

import (
	"fmt"

	"tellme/internal/config"
	"tellme/internal/llm"
	"tellme/internal/providers/anthropic"
	"tellme/internal/providers/anthropiccompat"
	"tellme/internal/providers/groq"
	"tellme/internal/providers/mistral"
	"tellme/internal/providers/openai"
	"tellme/internal/providers/openaicompat"
)

var opencodeGoMessagesModels = map[string]bool{
	"minimax-m3":   true,
	"minimax-m2.7": true,
	"minimax-m2.5": true,
	"qwen3.7-max":  true,
	"qwen3.7-plus": true,
	"qwen3.6-plus": true,
}

func New(cfg *config.Config) (llm.Provider, error) {
	inst := cfg.ActiveInstance()
	if inst == nil {
		return nil, fmt.Errorf("no active instance configured; run tellme --add-llm")
	}
	if inst.APIKey == "" {
		return nil, fmt.Errorf("instance %q has no API key", inst.Name)
	}

	switch inst.Provider {
	case "anthropic":
		return anthropic.New(inst.APIKey, inst.Model), nil
	case "openai":
		return openai.New(inst.APIKey, inst.Model), nil
	case "mistral":
		return mistral.New(inst.APIKey, inst.Model), nil
	case "groq":
		return groq.New(inst.APIKey, inst.Model), nil
	case "opencode-go":
		if opencodeGoMessagesModels[inst.Model] {
			return anthropiccompat.New(inst.APIKey, inst.Model, "https://opencode.ai/zen/go/v1/messages"), nil
		}
		return openaicompat.New(inst.APIKey, inst.Model, "https://opencode.ai/zen/go/v1/chat/completions", openaicompat.WithLowReasoning()), nil
	default:
		return nil, fmt.Errorf("unknown provider %q", inst.Provider)
	}
}
