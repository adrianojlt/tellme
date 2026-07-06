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

## Install

### Option 1: go install

From inside the project directory, install the binary into `~/go/bin`:

```bash
go install ./cmd/tellme
```

Make sure `~/go/bin` is in your `$PATH`. Add this to your `~/.zshrc` or `~/.bashrc` if it isn't already:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell and run from anywhere:

```bash
tellme "your question here"
```

### Option 2: Copy binary to /usr/local/bin

Build first, then copy:

```bash
go build -o tellme ./cmd/tellme
sudo cp tellme /usr/local/bin/tellme
```

On macOS, an unsigned Go binary copied into `/usr/local/bin` can be killed on launch with no error message (just `zsh: killed tellme`). macOS's `taskgated` rejects it. Clear any stale extended attributes, then apply an ad-hoc signature:

```bash
sudo xattr -cr /usr/local/bin/tellme
sudo codesign --force --sign - /usr/local/bin/tellme
```

If it still fails, check the architecture and the system log for the exact reason:

```bash
file /usr/local/bin/tellme
uname -m
log show --predicate 'process == "taskgated"' --last 1m
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
       tellme <subcommand>

Ask a natural language question and get shell command suggestions.

Example:
  tellme "list active docker containers"

Supported providers:
  anthropic    ANTHROPIC_API_KEY
  openai       OPENAI_API_KEY
  mistral      MISTRAL_API_KEY
  opencode-go  (API key from opencode.ai)

Config file: ~/.config/tellme/config.toml

Flags:
  --add-llm          Add or update an instance (interactive)
  --list-providers   List configured instances
  --set-provider     Switch the active instance
  --set-os           Set the target operating system
  --config           Show active configuration
  --help             Show this help message

Subcommands:
  history            Browse and re-run recent commands
  favorites          Browse and re-run saved favorites
  list               List all saved lists
  list <name>        Browse and re-run a named list
```

## Configuration

Run `tellme --add-llm` to interactively configure a provider instance. It will ask for a name, provider, model, and API key, then write `~/.config/tellme/config.toml`.

Use `--list-providers` to see all configured instances and `--set-provider` to switch between them. Use `--set-os` to tell tellme which OS to target when generating commands.

Example config with multiple instances:

```toml
active = "work-claude"
os = "macos"

[behavior]
max_options = 3
copy_after_select = false
max_history = 30

[[instances]]
name = "work-claude"
provider = "anthropic"
model = "claude-3-5-haiku-20241022"
api_key = "..."

[[instances]]
name = "personal-gpt"
provider = "openai"
model = "gpt-4o-mini"
api_key = "..."
```

API keys can also be set via environment variables (take precedence over the config file):

- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`
- `MISTRAL_API_KEY`

## History and Lists

After running a command, tellme saves it to history (capped at `max_history` entries).

```bash
tellme history       # browse recent commands, re-run or save one
tellme favorites     # browse saved favorites
tellme list          # show all named lists
tellme list <name>   # browse a named list
```

From the history view you can:
- Re-run or copy a command
- Add it to Favorites
- Add it to a named list (or create a new one)

From the favorites or list view you can:
- Re-run or copy a command
- Remove it from the list

## TODO

- Make the command available through `brew install tellme`
- Make the command available through `apt get install tellme`
