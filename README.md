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
       tellme [flags]

Ask a natural language question and get shell command suggestions.

Example:
  tellme "list active docker containers"

Supported providers:
  anthropic   ANTHROPIC_API_KEY
  openai      OPENAI_API_KEY
  mistral     MISTRAL_API_KEY
  groq        GROQ_API_KEY

Config file: ~/.config/tellme/config.toml

Flags:
  --add-llm        Add or update an LLM provider (interactive)
  --set-provider   Switch the active provider
  --config         Show active configuration
  --help           Show this help message
```

## Configuration

Run `tellme --add-llm` to interactively configure a provider - it will ask for the provider, model, and API key, then write `~/.config/tellme/config.toml` for you.

Alternatively, edit the file directly. The app reads `~/.config/tellme/config.toml` on startup. If the file is missing, defaults are used.

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