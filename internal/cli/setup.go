package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tellme/internal/config"
)

type modelInfo struct {
	displayName string
	id          string
}

type providerSetup struct {
	displayName  string
	id           string
	defaultModel string
	presetModels []modelInfo
}

var availableProviders = []providerSetup{
	{
		displayName:  "Anthropic",
		id:           "anthropic",
		defaultModel: "claude-haiku-4-5-20251001",
	},
	{
		displayName:  "OpenAI",
		id:           "openai",
		defaultModel: "gpt-4o-mini",
	},
	{
		displayName:  "Mistral",
		id:           "mistral",
		defaultModel: "mistral-small-latest",
		presetModels: []modelInfo{
			{"Mistral Small", "mistral-small-latest"},
		},
	},
	{
		displayName:  "OpenCode Go",
		id:           "opencode-go",
		defaultModel: "glm-5",
		presetModels: []modelInfo{
			{"GLM-5.1", "glm-5.1"},
			{"GLM-5", "glm-5"},
			{"Kimi K2.5", "kimi-k2.5"},
			{"Kimi K2.6", "kimi-k2.6"},
			{"DeepSeek V4 Pro", "deepseek-v4-pro"},
			{"DeepSeek V4 Flash", "deepseek-v4-flash"},
			{"MiMo-V2.5", "mimo-v2.5"},
			{"MiMo-V2.5-Pro", "mimo-v2.5-pro"},
			{"MiniMax M3", "minimax-m3"},
			{"MiniMax M2.7", "minimax-m2.7"},
			{"MiniMax M2.5", "minimax-m2.5"},
			{"Qwen3.7 Max", "qwen3.7-max"},
			{"Qwen3.7 Plus", "qwen3.7-plus"},
			{"Qwen3.6 Plus", "qwen3.6-plus"},
		},
	},
}

var availableOS = []string{"macos", "linux", "windows"}

func RunSetOS(cfgPath string) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Choose operating system:")
	for i, name := range availableOS {
		active := ""
		if name == cfg.OS {
			active = "  (active)"
		}
		fmt.Printf("  %d. %s%s\n", i+1, name, active)
	}

	for {

		fmt.Printf("Enter number (1-%d): ", len(availableOS))
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		n, parseErr := strconv.Atoi(strings.TrimSpace(line))
		if parseErr != nil || n < 1 || n > len(availableOS) {
			fmt.Printf("Invalid input. Enter a number between 1 and %d.\n", len(availableOS))
			continue
		}

		cfg.OS = availableOS[n-1]
		break
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("OS set to %s.\n", cfg.OS)
	return nil
}

func RunListProviders(cfgPath string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Instances) == 0 {
		fmt.Println("No instances configured. Run tellme --add-llm to add one.")
		return nil
	}

	fmt.Println("Configured instances:")
	for _, inst := range cfg.Instances {
		active := ""
		if inst.Name == cfg.Active {
			active = "  (active)"
		}
		fmt.Printf("  %-20s  %s / %s%s\n", inst.Name, inst.Provider, inst.Model, active)
	}

	return nil
}

func RunSetProvider(cfgPath string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Instances) == 0 {
		return fmt.Errorf("no instances configured; run tellme --add-llm first")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Available instances:")
	for i, inst := range cfg.Instances {
		active := ""
		if inst.Name == cfg.Active {
			active = "  (active)"
		}
		fmt.Printf("  %d. %s  (%s / %s)%s\n", i+1, inst.Name, inst.Provider, inst.Model, active)
	}

	var chosen config.Instance

	for {
		
		fmt.Printf("Enter number (1-%d): ", len(cfg.Instances))
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		n, parseErr := strconv.Atoi(strings.TrimSpace(line))
		if parseErr != nil || n < 1 || n > len(cfg.Instances) {
			fmt.Printf("Invalid input. Enter a number between 1 and %d.\n", len(cfg.Instances))
			continue
		}

		chosen = cfg.Instances[n-1]
		break
	}

	cfg.Active = chosen.Name

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Switched to %s.\n", chosen.Name)
	return nil
}

func RunAddLLM(cfgPath string) error {

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Which provider?")
	for i, p := range availableProviders {
		fmt.Printf("  %d. %s\n", i+1, p.displayName)
	}

	var chosen providerSetup
	for {

		fmt.Printf("Enter number (1-%d): ", len(availableProviders))
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		n, parseErr := strconv.Atoi(strings.TrimSpace(line))
		if parseErr != nil || n < 1 || n > len(availableProviders) {
			fmt.Printf("Invalid input. Enter a number between 1 and %d.\n", len(availableProviders))
			continue
		}

		chosen = availableProviders[n-1]
		break
	}

	model, err := pickModel(reader, chosen)
	if err != nil {
		return fmt.Errorf("picking model: %w", err)
	}

	fmt.Print("API key: ")
	keyLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	apiKey := strings.TrimSpace(keyLine)
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	inst := config.Instance{
		Name:     model,
		Provider: chosen.id,
		Model:    model,
		APIKey:   apiKey,
	}

	updated := false
	for i, existing := range cfg.Instances {
		if existing.Name == model {
			cfg.Instances[i] = inst
			updated = true
			break
		}
	}
	if !updated {
		cfg.Instances = append(cfg.Instances, inst)
	}

	cfg.Active = model

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Instance %q added and set as active.\n", model)
	return nil
}

func pickModel(reader *bufio.Reader, p providerSetup) (string, error) {

	if len(p.presetModels) == 0 {

		fmt.Printf("Model (e.g. %s): ", p.defaultModel)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}

		model := strings.TrimSpace(line)
		if model == "" {
			return p.defaultModel, nil
		}

		return model, nil
	}

	fmt.Println("Choose a model:")
	for i, m := range p.presetModels {
		fmt.Printf("  %d. %s (%s)\n", i+1, m.displayName, m.id)
	}
	fmt.Printf("  c. Custom model ID\n")

	for {

		fmt.Printf("Enter number or 'c': ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading input: %w", err)
		}

		input := strings.TrimSpace(line)

		if input == "c" {

			fmt.Printf("Model ID (e.g. %s): ", p.defaultModel)
			custom, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("reading input: %w", err)
			}

			model := strings.TrimSpace(custom)
			if model == "" {
				return p.defaultModel, nil
			}

			return model, nil
		}

		n, parseErr := strconv.Atoi(input)
		if parseErr != nil || n < 1 || n > len(p.presetModels) {
			fmt.Printf("Invalid input.\n")
			continue
		}

		return p.presetModels[n-1].id, nil
	}
}
