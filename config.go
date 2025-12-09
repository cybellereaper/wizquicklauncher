package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	return &config, nil
}

// ConfigGenerator represents the configuration generator UI
type ConfigGenerator struct {
	accounts []WizardInfo
	filePath string
	reader   *bufio.Reader
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
	account.Password = strings.TrimSpace(cg.prompt("Enter password: "))
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
	config := Config{
		FilePath:     cg.filePath,
		AccountsData: cg.accounts,
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile("config.json", data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("Configuration saved successfully!")
	return nil
}

// prompt reads a full line from stdin and trims the trailing newline.
func (cg *ConfigGenerator) prompt(message string) string {
	fmt.Print(message)
	input, _ := cg.reader.ReadString('\n')
	return strings.TrimSpace(input)
}
