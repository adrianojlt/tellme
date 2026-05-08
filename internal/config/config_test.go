package config

import (
	"os"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/tmp/nonexistent-tellme-config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.Provider != "openai" || cfg.Model != "gpt-4o-mini" || cfg.Behavior.MaxOptions != 3 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadMalformedTOML(t *testing.T) {
	f, err := os.CreateTemp("", "bad-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("not = valid [toml")
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "test-key")
	cfg, err := Load("/tmp/nonexistent-tellme-config.toml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Providers.Anthropic.APIKey != "test-key" {
		t.Fatalf("expected env override, got %q", cfg.Providers.Anthropic.APIKey)
	}
}
