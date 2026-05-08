package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type BehaviorConfig struct {
	MaxOptions      int  `toml:"max_options"`
	CopyAfterSelect bool `toml:"copy_after_select"`
}

type ProviderKeys struct {
	APIKey string `toml:"api_key"`
}

type ProvidersConfig struct {
	OpenAI    ProviderKeys `toml:"openai"`
	Anthropic ProviderKeys `toml:"anthropic"`
	Mistral   ProviderKeys `toml:"mistral"`
	Groq      ProviderKeys `toml:"groq"`
}

type Config struct {
	Provider  string          `toml:"provider"`
	Model     string          `toml:"model"`
	Behavior  BehaviorConfig  `toml:"behavior"`
	Providers ProvidersConfig `toml:"providers"`
}

func Default() *Config {
	return &Config{
		Provider: "openai",
		Model:    "gpt-4o-mini",
		Behavior: BehaviorConfig{
			MaxOptions:      3,
			CopyAfterSelect: false,
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			applyEnvOverrides(cfg)
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	if _, err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.Providers.OpenAI.APIKey = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.Providers.Anthropic.APIKey = v
	}
	if v := os.Getenv("MISTRAL_API_KEY"); v != "" {
		cfg.Providers.Mistral.APIKey = v
	}
	if v := os.Getenv("GROQ_API_KEY"); v != "" {
		cfg.Providers.Groq.APIKey = v
	}
}
