package mistral

import "tellme/internal/providers/openaicompat"

func New(apiKey, model string) *openaicompat.Client {
	return openaicompat.New(apiKey, model, "https://api.mistral.ai/v1/chat/completions")
}
