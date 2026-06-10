package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type BehaviorConfig struct {
	MaxOptions      int  `toml:"max_options"`
	CopyAfterSelect bool `toml:"copy_after_select"`
}

type Instance struct {
	Name     string `toml:"name"`
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	APIKey   string `toml:"api_key"`
}

type Config struct {
	Active    string         `toml:"active"`
	OS        string         `toml:"os"`
	Behavior  BehaviorConfig `toml:"behavior"`
	Instances []Instance     `toml:"instances"`
}

func (c *Config) ActiveInstance() *Instance {
	for i := range c.Instances {
		if c.Instances[i].Name == c.Active {
			return &c.Instances[i]
		}
	}
	return nil
}

func Default() *Config {
	return &Config{
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
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()

	if _, err := toml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}
