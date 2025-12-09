package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	expected := Config{
		FilePath: "C:/Games/Wizard101",
		AccountsData: []WizardInfo{{
			Username: "test-user",
			Password: "secret",
			XPos:     100,
			YPos:     200,
		}},
	}

	data, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.FilePath != expected.FilePath {
		t.Fatalf("unexpected FilePath: got %q, want %q", config.FilePath, expected.FilePath)
	}

	if len(config.AccountsData) != len(expected.AccountsData) {
		t.Fatalf("unexpected accounts length: got %d, want %d", len(config.AccountsData), len(expected.AccountsData))
	}

	if config.AccountsData[0].Username != expected.AccountsData[0].Username {
		t.Fatalf("unexpected username: got %q, want %q", config.AccountsData[0].Username, expected.AccountsData[0].Username)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/path/does/not/exist.json")
	if err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
