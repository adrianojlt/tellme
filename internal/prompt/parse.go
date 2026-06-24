package prompt

import (
	"encoding/json"
	"errors"
	"fmt"

	"tellme/internal/domain"
)

var ErrNoSuggestions = errors.New("no suggestions in response")

type LLMResponse struct {
	Suggestions []domain.CommandSuggestion `json:"suggestions"`
}

func ParseResponse(raw string) ([]domain.CommandSuggestion, error) {

	var resp LLMResponse

	if err := json.Unmarshal([]byte(raw), &resp); err != nil {

		snippet := raw

		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}

		return nil, fmt.Errorf("invalid JSON response: %w (got: %s)", err, snippet)
	}

	if len(resp.Suggestions) == 0 {
		return nil, ErrNoSuggestions
	}

	return resp.Suggestions, nil
}
