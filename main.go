package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// WizardInfo represents the information for a single Wizard101 account
type WizardInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	XPos     int    `json:"x"`
	YPos     int    `json:"y"`
	Status   string `json:"-"`
}

// Config represents the configuration options for the program
type Config struct {
	FilePath     string       `json:"filePath"`
	AccountsData []WizardInfo `json:"accountsData"`
}

// Application represents the main application
type Application struct {
	Config     *Config
	wm         *WindowManager
	processes  []*os.Process
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewApplication creates a new application instance
func NewApplication(config *Config) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		Config:     config,
		wm:         NewWindowManager(),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// OpenWizard opens the Wizard101 application and returns the process
func (app *Application) OpenWizard() (*os.Process, error) {
	cmd := exec.Command(
		"cmd", "/C",
		"cd", app.Config.FilePath, "&&",
		"start", "WizardGraphicalClient.exe",
		"-L", "login.us.wizard101.com", "12000",
	)
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Wizard101: %w", err)
	}
	return cmd.Process, nil
}

// LaunchAccounts launches all accounts in the configuration
func (app *Application) LaunchAccounts() error {
	initialHandles := app.wm.GetAllWizardHandles()
	fmt.Println("Starting Wizard101 Quick Launcher by CybelleReaper")
	fmt.Printf("Found %d existing Wizard101 windows\n", len(initialHandles))

	// Launch all clients
	for _, account := range app.Config.AccountsData {
		fmt.Printf("Launching account: %s...\n", account.Username)

		process, err := app.OpenWizard()
		if err != nil {
			fmt.Printf("Failed to launch account %s: %v\n", account.Username, err)
			continue
		}
		app.processes = append(app.processes, process)
	}

	fmt.Println("Waiting for windows to open...")
	time.Sleep(2 * time.Second)

	// Track new handles
	newHandles := make(map[windows.Handle]struct{})
	for _, account := range app.Config.AccountsData {
		select {
		case <-app.ctx.Done():
			return fmt.Errorf("operation cancelled")
		default:
			handle := app.waitForNextWindow(initialHandles, newHandles)
			if handle == 0 {
				fmt.Printf("Failed to detect window for account %s\n", account.Username)
				continue
			}

			fmt.Printf("Logging in account: %s\n", account.Username)
			app.wm.WizardLogin(handle, account.Username, account.Password)
			app.wm.MoveWindow(handle, account.XPos, account.YPos)
			fmt.Printf("Account %s successfully logged in and positioned\n", account.Username)
		}
	}

	fmt.Println("All accounts processed. Press Ctrl+C to exit.")
	<-app.ctx.Done()
	return nil
}

// waitForNextWindow waits for a new Wizard101 window to appear
func (app *Application) waitForNextWindow(initialHandles, newHandles map[windows.Handle]struct{}) windows.Handle {
	fmt.Println("Waiting for next window...")
	for attempt := 0; attempt < 60; attempt++ {
		handles := app.wm.GetAllWizardHandles()
		for handle := range handles {
			if _, exists := initialHandles[handle]; !exists {
				if _, exists := newHandles[handle]; !exists {
					newHandles[handle] = struct{}{}
					return handle
				}
			}
		}
		select {
		case <-app.ctx.Done():
			return 0
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
	fmt.Println("Timed out waiting for window")
	return 0
}

// Run runs the main application logic
func (app *Application) Run() error {
	fmt.Println("Wizard101 Quick Launcher")
	fmt.Println("------------------------")
	fmt.Printf("Loaded %d accounts from configuration\n", len(app.Config.AccountsData))

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancelFunc = cancel

	// Handle Ctrl+C
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\nShutting down...")
		cancel()
	}()

	return app.LaunchAccounts()
}

func main() {
	var config *Config
	var err error

	// Try to load existing config first
	config, err = LoadConfig("config.json")

	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, run the generator
			configGen := NewConfigGenerator()
			if err := configGen.RunUI(); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating config: %v\n", err)
				os.Exit(1)
			}
			// Load the newly generated config
			config, err = LoadConfig("config.json")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}

	app := NewApplication(config)
	defer func() {
		app.cancelFunc()
		for _, p := range app.processes {
			p.Release()
		}
	}()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
