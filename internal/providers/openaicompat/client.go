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
	baseURL string
	apiKey  string
	model   string
	http    *http.Client
}

func New(apiKey, model, baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		http:    &http.Client{},
	}
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type request struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	ResponseFormat responseFormat `json:"response_format"`
}

type choice struct {
	Message message `json:"message"`
}

type response struct {
	Choices []choice `json:"choices"`
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (c *Client) SuggestCommands(ctx context.Context, req domain.SuggestRequest) ([]domain.CommandSuggestion, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)

	defer cancel()

	body, err := json.Marshal(request{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: prompt.SystemPrompt()},
			{Role: "user", Content: prompt.UserPrompt(req)},
		},
		ResponseFormat: responseFormat{Type: "json_object"},
	})

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

		return nil, fmt.Errorf("API error (%s): %s", apiErr.Error.Type, apiErr.Error.Message)
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
