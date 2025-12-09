//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/windows"
)

// Application represents the main application lifecycle and orchestration.
type Application struct {
	Config     *Config
	wm         *WindowManager
	processes  []*os.Process
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewApplication creates a new application instance.
func NewApplication(config *Config) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		Config:     config,
		wm:         NewWindowManager(),
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

// OpenWizard opens the Wizard101 application and returns the process.
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

// LaunchAccounts launches all accounts in the configuration.
func (app *Application) LaunchAccounts() error {
	const (
		windowPollInterval = 500 * time.Millisecond
		windowPollAttempts = 60
	)

	initialHandles := app.wm.GetAllWizardHandles()
	fmt.Println("Starting Wizard101 Quick Launcher by CybelleReaper")
	fmt.Printf("Found %d existing Wizard101 windows\n", len(initialHandles))

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
	newHandles := make(map[windows.Handle]struct{})

	for _, account := range app.Config.AccountsData {
		select {
		case <-app.ctx.Done():
			return fmt.Errorf("operation cancelled")
		default:
			handle := app.waitForNextWindow(initialHandles, newHandles, windowPollAttempts, windowPollInterval)
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

// waitForNextWindow waits for a new Wizard101 window to appear.
func (app *Application) waitForNextWindow(initialHandles, newHandles map[windows.Handle]struct{}, attempts int, delay time.Duration) windows.Handle {
	fmt.Println("Waiting for next window...")
	for attempt := 0; attempt < attempts; attempt++ {
		handles := app.wm.GetAllWizardHandles()
		for handle := range handles {
			if _, exists := initialHandles[handle]; exists {
				continue
			}
			if _, exists := newHandles[handle]; exists {
				continue
			}

			newHandles[handle] = struct{}{}
			return handle
		}

		select {
		case <-app.ctx.Done():
			return 0
		default:
			time.Sleep(delay)
		}
	}

	fmt.Println("Timed out waiting for window")
	return 0
}

// Run runs the main application logic.
func (app *Application) Run() error {
	fmt.Println("Wizard101 Quick Launcher")
	fmt.Println("------------------------")
	fmt.Printf("Loaded %d accounts from configuration\n", len(app.Config.AccountsData))

	return app.LaunchAccounts()
}
