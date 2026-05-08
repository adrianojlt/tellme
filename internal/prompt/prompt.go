package prompt

import (
	"fmt"

	"tellme/internal/domain"
)

func SystemPrompt() string {
	return `You are a shell command assistant. Respond with JSON only. No markdown, no code fences, no explanation.

Use exactly this schema:
{
  "suggestions": [
    {
      "title": "short title",
      "command": "the shell command",
      "description": "what it does",
      "risk_level": "low|medium|high",
      "requires_sudo": false,
      "platform": "linux|macos|windows|any"
    }
  ]
}`
}

func UserPrompt(req domain.SuggestRequest) string {

	msg := fmt.Sprintf("Query: %s", req.UserQuery)

	if req.MaxOptions > 0 {
		msg += fmt.Sprintf("\nProvide %d suggestions.", req.MaxOptions)
	}

	if req.OSHint != "" {
		msg += fmt.Sprintf("\nOS: %s", req.OSHint)
	}

	if req.ShellHint != "" {
		msg += fmt.Sprintf("\nShell: %s", req.ShellHint)
	}

	return msg
}
