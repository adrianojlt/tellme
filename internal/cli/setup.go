package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tellme/internal/config"
)

type providerInfo struct {
	name         string
	configKey    string
	defaultModel string
}

var providers = []providerInfo{
	{"Anthropic", "anthropic", "claude-haiku-4-5-20251001"},
	{"OpenAI", "openai", "gpt-4o-mini"},
	{"Mistral", "mistral", "mistral-small-latest"},
	{"Groq", "groq", "llama3-8b-8192"},
}

func apiKeyFor(cfg *config.Config, key string) string {
	switch key {
	case "anthropic":
		return cfg.Providers.Anthropic.APIKey
	case "openai":
		return cfg.Providers.OpenAI.APIKey
	case "mistral":
		return cfg.Providers.Mistral.APIKey
	case "groq":
		return cfg.Providers.Groq.APIKey
	}
	return ""
}

func RunSetProvider(cfgPath string) error {
	cfg, err := config.LoadRaw(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var configured []providerInfo
	for _, p := range providers {
		if apiKeyFor(cfg, p.configKey) != "" {
			configured = append(configured, p)
		}
	}
	if len(configured) == 0 {
		return fmt.Errorf("no providers configured - run tellme --add-llm first")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Configured providers:")
	for i, p := range configured {
		active := ""
		if p.configKey == cfg.Provider {
			active = "  (active)"
		}
		fmt.Printf("  %d. %s%s\n", i+1, p.name, active)
	}

	var chosen providerInfo
	for {
		fmt.Printf("Enter number (1-%d): ", len(configured))
		line, _ := reader.ReadString('\n')
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 1 || n > len(configured) {
			fmt.Printf("Invalid input. Enter a number between 1 and %d.\n", len(configured))
			continue
		}
		chosen = configured[n-1]
		break
	}

	fmt.Printf("Model (e.g. %s) [current: %s]: ", chosen.defaultModel, cfg.Model)
	modelLine, _ := reader.ReadString('\n')
	model := strings.TrimSpace(modelLine)
	if model == "" {
		model = cfg.Model
	}

	cfg.Provider = chosen.configKey
	cfg.Model = model

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Switched to %s (%s).\n", chosen.name, model)
	return nil
}

func RunAddLLM(cfgPath string) error {
	cfg, err := config.LoadRaw(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Which provider?")
	for i, p := range providers {
		fmt.Printf("  %d. %s\n", i+1, p.name)
	}

	var chosen providerInfo
	for {
		fmt.Printf("Enter number (1-%d): ", len(providers))
		line, _ := reader.ReadString('\n')
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 1 || n > len(providers) {
			fmt.Printf("Invalid input. Enter a number between 1 and %d.\n", len(providers))
			continue
		}
		chosen = providers[n-1]
		break
	}

	fmt.Printf("Model (e.g. %s): ", chosen.defaultModel)
	modelLine, _ := reader.ReadString('\n')
	model := strings.TrimSpace(modelLine)
	if model == "" {
		model = chosen.defaultModel
	}

	fmt.Print("API key: ")
	keyLine, _ := reader.ReadString('\n')
	apiKey := strings.TrimSpace(keyLine)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	cfg.Provider = chosen.configKey
	cfg.Model = model

	switch chosen.configKey {
	case "anthropic":
		cfg.Providers.Anthropic.APIKey = apiKey
	case "openai":
		cfg.Providers.OpenAI.APIKey = apiKey
	case "mistral":
		cfg.Providers.Mistral.APIKey = apiKey
	case "groq":
		cfg.Providers.Groq.APIKey = apiKey
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Provider %s configured and saved.\n", chosen.name)
	return nil
}
