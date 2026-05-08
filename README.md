# tellme

Ask a natural language question and get shell command suggestions from an LLM.

```bash
tellme "list active docker containers"
```

## Prerequisites

- Go 1.21 or later

## Build

```bash
go build -o tellme ./cmd/tellme
```

## Run

```bash
# From source
go run ./cmd/tellme "your question here"

# From binary
./tellme "your question here"
```

## Usage

```
Usage: tellme <query>

Ask a natural language question and get shell command suggestions.

Example:
  tellme "list active docker containers"
```

## Configuration

Configuration support is coming in Part 5. The app will read from `~/.config/tellme/config.toml` and support:

- LLM provider selection (openai, anthropic, mistral, groq)
- Model selection
- API keys (also via environment variables)
- Max number of suggestions
- Auto-copy to clipboard

## Current state

Part 4 complete: the app accepts a query, calls the fake provider, displays a numbered list of suggestions, and prompts the user to pick one by number. Entering 0 exits cleanly; a valid number prints the selected command. Invalid input re-prompts. Real LLM integration is coming next.
