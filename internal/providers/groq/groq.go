package groq

import "tellme/internal/providers/openaicompat"

func New(apiKey, model string) *openaicompat.Client {
	return openaicompat.New(apiKey, model, "https://api.groq.com/openai/v1/chat/completions")
}
