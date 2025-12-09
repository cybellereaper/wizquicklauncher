package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// LoadConfig loads the configuration from a file.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if !config.UsesEncryption {
		for i := range config.AccountsData {
			config.AccountsData[i].Password = config.AccountsData[i].EncryptedPassword
		}
		return &config, nil
	}

	passphrase := strings.TrimSpace(os.Getenv(passphraseEnvVar))
	if passphrase == "" {
		return nil, fmt.Errorf("configuration requires encryption passphrase: set %s", passphraseEnvVar)
	}

	salt, err := base64.StdEncoding.DecodeString(config.EncryptionSalt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption salt: %w", err)
	}

	for i := range config.AccountsData {
		plain, decErr := decryptSecret(config.AccountsData[i].EncryptedPassword, passphrase, salt)
		if decErr != nil {
			return nil, fmt.Errorf("failed to decrypt password for account %s: %w", config.AccountsData[i].Username, decErr)
		}
		config.AccountsData[i].Password = plain
	}
	return &config, nil
}

// ConfigGenerator represents the configuration generator UI
type ConfigGenerator struct {
	accounts   []WizardInfo
	filePath   string
	reader     *bufio.Reader
	passphrase string
	salt       []byte
}

// NewConfigGenerator creates a new configuration generator instance
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{
		accounts: make([]WizardInfo, 0),
		filePath: "",
		reader:   bufio.NewReader(os.Stdin),
	}
}

// RunUI runs the configuration generator user interface
func (cg *ConfigGenerator) RunUI() error {
	fmt.Println("Wizard101 Quick Launcher - Configuration Generator")
	fmt.Println("----------------------------------------------")

	// Get Wizard101 installation path
	cg.filePath = cg.prompt("Enter Wizard101 installation path: ")
	cg.passphrase = strings.TrimSpace(os.Getenv(passphraseEnvVar))
	if cg.passphrase == "" {
		fmt.Println("\nCreate a secure passphrase to protect your saved passwords.")
		fmt.Printf("The same passphrase will be required via the %s environment variable when launching.\n", passphraseEnvVar)
		if err := cg.promptForPassphrase(); err != nil {
			return err
		}
	}

	if len(cg.salt) == 0 {
		newSalt, err := generateSalt()
		if err != nil {
			return err
		}
		cg.salt = newSalt
	}

	for {
		fmt.Println("\nCurrent accounts:", len(cg.accounts))
		fmt.Println("1. Add account")
		fmt.Println("2. List accounts")
		fmt.Println("3. Save configuration")
		fmt.Println("4. Exit")

		choice := cg.prompt("Choose an option (1-4): ")

		switch strings.TrimSpace(choice) {
		case "1":
			cg.addAccount()
		case "2":
			cg.listAccounts()
		case "3":
			if err := cg.saveConfig(); err != nil {
				return err
			}
		case "4":
			return nil
		default:
			fmt.Println("Invalid option")
		}
	}
}

// addAccount adds a new account to the configuration
func (cg *ConfigGenerator) addAccount() {
	var account WizardInfo

	account.Username = strings.TrimSpace(cg.prompt("Enter username: "))
	password, err := cg.promptSecret("Enter password: ")
	if err != nil {
		fmt.Printf("Failed to read password: %v\n", err)
		return
	}

	account.Password = strings.TrimSpace(password)
	fmt.Sscanf(cg.prompt("Enter X position: "), "%d", &account.XPos)
	fmt.Sscanf(cg.prompt("Enter Y position: "), "%d", &account.YPos)

	cg.accounts = append(cg.accounts, account)
	fmt.Println("Account added successfully!")
}

// listAccounts displays all configured accounts
func (cg *ConfigGenerator) listAccounts() {
	if len(cg.accounts) == 0 {
		fmt.Println("No accounts configured")
		return
	}

	fmt.Println("\nConfigured accounts:")
	for i, acc := range cg.accounts {
		fmt.Printf("%d. Username: %s, Position: (%d, %d)\n",
			i+1, acc.Username, acc.XPos, acc.YPos)
	}
}

// saveConfig saves the configuration to a JSON file
func (cg *ConfigGenerator) saveConfig() error {
	if cg.passphrase == "" {
		return errors.New("a passphrase is required to save the configuration")
	}

	if len(cg.salt) == 0 {
		newSalt, err := generateSalt()
		if err != nil {
			return err
		}
		cg.salt = newSalt
	}

	config := Config{
		FilePath:       cg.filePath,
		AccountsData:   make([]WizardInfo, 0, len(cg.accounts)),
		UsesEncryption: true,
		EncryptionSalt: base64.StdEncoding.EncodeToString(cg.salt),
	}

	for _, account := range cg.accounts {
		encrypted, err := encryptSecret(account.Password, cg.passphrase, cg.salt)
		if err != nil {
			return fmt.Errorf("failed to encrypt password for %s: %w", account.Username, err)
		}

		config.AccountsData = append(config.AccountsData, WizardInfo{
			Username:          account.Username,
			EncryptedPassword: encrypted,
			XPos:              account.XPos,
			YPos:              account.YPos,
		})
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("config.json", data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("Configuration saved successfully!")
	return nil
}

// promptSecret securely reads a secret value (like a password) without echoing it back to the terminal.
func (cg *ConfigGenerator) promptSecret(message string) (string, error) {
	fmt.Print(message)
	secret, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	return string(secret), nil
}

// promptForPassphrase asks the user to create and confirm a passphrase for encrypting credentials.
func (cg *ConfigGenerator) promptForPassphrase() error {
	const minLength = 12

	for attempts := 0; attempts < 3; attempts++ {
		candidate, err := cg.promptSecret("Create passphrase (min 12 characters): ")
		if err != nil {
			return err
		}

		if len(candidate) < minLength {
			fmt.Printf("Passphrase must be at least %d characters.\n", minLength)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		confirm, err := cg.promptSecret("Confirm passphrase: ")
		if err != nil {
			return err
		}

		if candidate != confirm {
			fmt.Println("Passphrases do not match.")
			time.Sleep(500 * time.Millisecond)
			continue
		}

		cg.passphrase = candidate
		return nil
	}

	return errors.New("failed to set passphrase after multiple attempts")
}

// prompt reads a full line from stdin and trims the trailing newline.
func (cg *ConfigGenerator) prompt(message string) string {
	fmt.Print(message)
	input, _ := cg.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
