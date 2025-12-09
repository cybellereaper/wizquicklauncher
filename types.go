package main

// WizardInfo represents the information for a single Wizard101 account.
// Kept in a standalone file so it is available to both Windows-only and
// platform-neutral code (for example, tests and configuration helpers).
type WizardInfo struct {
	Username          string `json:"username"`
	EncryptedPassword string `json:"password"`
	Password          string `json:"-"`
	XPos              int    `json:"x"`
	YPos              int    `json:"y"`
	Status            string `json:"-"`
}

// Config represents the configuration options for the program.
type Config struct {
	FilePath       string       `json:"filePath"`
	AccountsData   []WizardInfo `json:"accountsData"`
	UsesEncryption bool         `json:"usesEncryption"`
	EncryptionSalt string       `json:"encryptionSalt,omitempty"`
}
