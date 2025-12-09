package main

import (
	"encoding/base64"
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
			Username:          "test-user",
			EncryptedPassword: "secret",
			XPos:              100,
			YPos:              200,
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

	if config.AccountsData[0].Password != "secret" {
		t.Fatalf("unexpected password: got %q", config.AccountsData[0].Password)
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

func TestLoadConfig_WithEncryption(t *testing.T) {
	passphrase := "super-secure-passphrase"
	salt, err := generateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	encryptedPassword, err := encryptSecret("hunter2", passphrase, salt)
	if err != nil {
		t.Fatalf("failed to encrypt password: %v", err)
	}

	config := Config{
		FilePath:       "C:/Games/Wizard101",
		UsesEncryption: true,
		EncryptionSalt: base64.StdEncoding.EncodeToString(salt),
		AccountsData: []WizardInfo{{
			Username:          "encrypted-user",
			EncryptedPassword: encryptedPassword,
			XPos:              10,
			YPos:              20,
		}},
	}

	configPath := filepath.Join(t.TempDir(), "config.json")
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal encrypted config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("failed to write encrypted config: %v", err)
	}

	t.Setenv(passphraseEnvVar, passphrase)
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load encrypted config: %v", err)
	}

	if loaded.AccountsData[0].Password != "hunter2" {
		t.Fatalf("expected decrypted password to match, got %q", loaded.AccountsData[0].Password)
	}
}
