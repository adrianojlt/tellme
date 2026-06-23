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

func TestDefaultMaxHistory(t *testing.T) {
	cfg := Default()
	if cfg.Behavior.MaxHistory != 30 {
		t.Fatalf("expected default MaxHistory 30, got %d", cfg.Behavior.MaxHistory)
	}
}

func TestLoadMaxHistoryOmitted(t *testing.T) {
	f, err := os.CreateTemp("", "omit-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("[behavior]\nmax_options = 5\n")
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Behavior.MaxHistory != 30 {
		t.Fatalf("expected MaxHistory 30 when omitted, got %d", cfg.Behavior.MaxHistory)
	}
}

func TestLoadMaxHistoryExplicit(t *testing.T) {
	f, err := os.CreateTemp("", "explicit-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("[behavior]\nmax_history = 50\n")
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Behavior.MaxHistory != 50 {
		t.Fatalf("expected MaxHistory 50, got %d", cfg.Behavior.MaxHistory)
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
