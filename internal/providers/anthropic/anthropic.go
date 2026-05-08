package anthropic

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

const apiURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

func New(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type response struct {
	Content []contentBlock `json:"content"`
}

type apiError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) SuggestCommands(ctx context.Context, req domain.SuggestRequest) ([]domain.CommandSuggestion, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	body, err := json.Marshal(request{
		Model:     c.model,
		MaxTokens: 1024,
		System:    prompt.SystemPrompt(),
		Messages:  []message{{Role: "user", Content: prompt.UserPrompt(req)}},
	})
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("anthropic error (HTTP %d)", resp.StatusCode)
		}
		return nil, fmt.Errorf("anthropic error (%s): %s", apiErr.Error.Type, apiErr.Error.Message)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from anthropic")
	}

	return prompt.ParseResponse(result.Content[0].Text)
}
