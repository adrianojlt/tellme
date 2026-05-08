package openai

import "tellme/internal/providers/openaicompat"

func New(apiKey, model string) *openaicompat.Client {
	return openaicompat.New(apiKey, model, "https://api.openai.com/v1/chat/completions")
}
