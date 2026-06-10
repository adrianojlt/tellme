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
	if cfg.Behavior.MaxOptions != 3 {
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

func TestActiveInstance(t *testing.T) {
	cfg := &Config{
		Active: "gpt-4o-mini",
		Instances: []Instance{
			{Name: "gpt-4o-mini", Provider: "openai", Model: "gpt-4o-mini", APIKey: "k1"},
			{Name: "glm-5", Provider: "opencode-go", Model: "glm-5", APIKey: "k2"},
		},
	}

	inst := cfg.ActiveInstance()
	if inst == nil || inst.Provider != "openai" {
		t.Fatalf("expected openai instance, got %v", inst)
	}

	cfg.Active = "glm-5"
	inst = cfg.ActiveInstance()
	if inst == nil || inst.Provider != "opencode-go" {
		t.Fatalf("expected opencode-go instance, got %v", inst)
	}

	cfg.Active = "nonexistent"
	if cfg.ActiveInstance() != nil {
		t.Fatal("expected nil for unknown active")
	}
}
