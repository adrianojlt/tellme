package openaicompat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tellme/internal/domain"
	"tellme/internal/prompt"
)

type Client struct {
	baseURL         string
	apiKey          string
	model           string
	reasoningEffort string
	disableThinking bool
	http            *http.Client
}

func New(apiKey, model, baseURL string, opts ...func(*Client)) *Client {
	c := &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func WithLowReasoning() func(*Client) {
	return func(c *Client) {
		c.reasoningEffort = "low"
		c.disableThinking = true
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type thinkingConfig struct {
	Type string `json:"type"`
}

type request struct {
	Model           string          `json:"model"`
	MaxTokens       int             `json:"max_tokens"`
	Messages        []message       `json:"messages"`
	ResponseFormat  responseFormat  `json:"response_format"`
	ReasoningEffort string          `json:"reasoning_effort,omitempty"`
	Thinking        *thinkingConfig `json:"thinking,omitempty"`
}

type choice struct {
	Message message `json:"message"`
}

type response struct {
	Choices []choice `json:"choices"`
}

type apiError struct {
	// OpenAI/Groq format: {"error": {"message": "...", "type": "..."}}
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	// Mistral format: {"message": "...", "type": "..."}
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (c *Client) SuggestCommands(ctx context.Context, req domain.SuggestRequest) ([]domain.CommandSuggestion, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

	defer cancel()

	r := request{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []message{
			{Role: "system", Content: prompt.SystemPrompt()},
			{Role: "user", Content: prompt.UserPrompt(req)},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
	}

	if c.reasoningEffort != "" {
		r.ReasoningEffort = c.reasoningEffort
	}

	if c.disableThinking {
		r.Thinking = &thinkingConfig{Type: "disabled"}
	}

	body, err := json.Marshal(r)

	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("API error (HTTP %d)", resp.StatusCode)
		}

		errType := apiErr.Error.Type
		if errType == "" {
			errType = apiErr.Type
		}

		errMsg := apiErr.Error.Message
		if errMsg == "" {
			errMsg = apiErr.Message
		}

		return nil, fmt.Errorf("API error (%s): %s", errType, errMsg)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	return prompt.ParseResponse(result.Choices[0].Message.Content)
}
