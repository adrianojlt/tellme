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

The app reads `~/.config/tellme/config.toml` on startup. If the file is missing, defaults are used.

```toml
provider = "openai"          # openai | anthropic | mistral | groq
model = "gpt-4o-mini"

[behavior]
max_options = 3
copy_after_select = false

[providers.openai]
api_key = "..."

[providers.anthropic]
api_key = "..."
```

API keys can also be set via environment variables (take precedence over the file):

- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `MISTRAL_API_KEY`
- `GROQ_API_KEY`

## Current state

Part 7 complete: the app calls the Anthropic Messages API to generate real shell command suggestions. Set `provider = "anthropic"` in the config (or use the env var `ANTHROPIC_API_KEY`) and run with any query. OpenAI, Mistral, and Groq support is coming in Part 9.
