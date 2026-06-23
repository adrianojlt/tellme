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
	storePath := filepath.Join(homeDir, ".config", "tellme", "store.json")

	if len(args) == 1 && args[0] == "--add-llm" {
		if err := cli.RunAddLLM(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(args) == 1 && args[0] == "--set-os" {
		if err := cli.RunSetOS(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(args) == 1 && args[0] == "--list-providers" {
		if err := cli.RunListProviders(cfgPath); err != nil {
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

		inst := cfg.ActiveInstance()
		fmt.Printf("Active instance: %s\n", cfg.Active)
		if inst != nil {
			fmt.Printf("  Provider: %s\n", inst.Provider)
			fmt.Printf("  Model:    %s\n", inst.Model)
		} else {
			fmt.Printf("  (none configured)\n")
		}
		osDisplay := cfg.OS
		if osDisplay == "" {
			osDisplay = "(not set)"
		}
		fmt.Printf("OS: %s\n", osDisplay)
		fmt.Printf("\nConfigured instances: %d\n", len(cfg.Instances))
		for _, i := range cfg.Instances {
			active := ""
			if i.Name == cfg.Active {
				active = "  *"
			}
			fmt.Printf("  %-20s  %s / %s%s\n", i.Name, i.Provider, i.Model, active)
		}

		os.Exit(0)
	}

	if cli.IsSubcommand(args) {

		var err error

		switch args[0] {
		case "history":
			err = cli.RunHistory(cfgPath)
		case "favorites":
			err = cli.RunFavorites(cfgPath)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}

	// List subcommands: `tellme list` (names) and `tellme list <name>`. Only
	// arity 1 and 2 are subcommands; a 3+ token `list ...` falls through to
	// query handling. In practice a query arrives as one quoted arg, so the
	// two-token `tellme list foo` is treated as the named-list subcommand by
	// design.
	if cli.IsListNames(args) {
		if err := cli.RunListNames(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if cli.IsListNamed(args) {
		if err := cli.RunList(cfgPath, args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(args) == 1 && !strings.Contains(args[0], " ") {
		fmt.Fprintf(os.Stderr, "Unknown argument: %s\n", args[0])
		fmt.Fprintf(os.Stderr, "Run 'tellme --help' for usage.\n")
		os.Exit(1)
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

	a := app.New(p, cli.PrintSuggestions, cli.SelectSuggestion, cfg.Behavior.MaxOptions, clipboard.Copy, cfg.Behavior.CopyAfterSelect, cfg.OS, storePath)
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
