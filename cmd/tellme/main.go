package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tellme/internal/app"
	"tellme/internal/cli"
	"tellme/internal/clipboard"
	"tellme/internal/config"
	"tellme/internal/prompt"
	"tellme/internal/providers"
)

const usage = `Usage: tellme <query>
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
`

func main() {
	args := os.Args[1:]

	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Print(usage)
		os.Exit(0)
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cfgPath := filepath.Join(homeDir, ".config", "tellme", "config.toml")

	if len(args) == 1 && args[0] == "--add-llm" {

		if err := cli.RunAddLLM(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if len(args) == 1 && args[0] == "--set-provider" {

		if err := cli.RunSetProvider(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if len(args) == 1 && args[0] == "--config" {

		cfg, err := config.Load(cfgPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		keyStatus := func(k string) string {
			if k != "" {
				return "set"
			}
			return "not set"
		}

		fmt.Printf("Active configuration\n")
		fmt.Printf("  Provider:  %s\n", cfg.Provider)
		fmt.Printf("  Model:     %s\n\n", cfg.Model)
		fmt.Printf("API keys\n")
		fmt.Printf("  Anthropic: %s\n", keyStatus(cfg.Providers.Anthropic.APIKey))
		fmt.Printf("  OpenAI:    %s\n", keyStatus(cfg.Providers.OpenAI.APIKey))
		fmt.Printf("  Mistral:   %s\n", keyStatus(cfg.Providers.Mistral.APIKey))
		fmt.Printf("  Groq:      %s\n", keyStatus(cfg.Providers.Groq.APIKey))
		
		os.Exit(0)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	p, err := providers.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	query := args[0]

	if strings.TrimSpace(query) == "" {
		fmt.Fprintln(os.Stderr, "Error: query cannot be empty")
		os.Exit(1)
	}

	a := app.New(p, cli.PrintSuggestions, cli.SelectSuggestion, cfg.Behavior.MaxOptions, clipboard.Copy, cfg.Behavior.CopyAfterSelect)
	if err := a.Run(context.Background(), query); err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			fmt.Fprintln(os.Stderr, "Request timed out. Check your connection or try again.")
		case errors.Is(err, prompt.ErrNoSuggestions):
			fmt.Fprintln(os.Stderr, "No suggestions returned. Try rephrasing your query.")
		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
