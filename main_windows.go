//go:build windows
// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config, err := loadOrGenerateConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing configuration: %v\n", err)
		os.Exit(1)
	}

	app := NewApplication(config)
	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx
	app.cancelFunc = cancel

	defer func() {
		app.cancelFunc()
		for _, p := range app.processes {
			_ = p.Release()
		}
	}()

	go handleSignals(cancel)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSignals(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")
	cancel()
}

func loadOrGenerateConfig() (*Config, error) {
	config, err := LoadConfig("config.json")
	if err == nil {
		return config, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	configGen := NewConfigGenerator()
	if genErr := configGen.RunUI(); genErr != nil {
		return nil, fmt.Errorf("error generating config: %w", genErr)
	}

	config, err = LoadConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("error loading generated config: %w", err)
	}

	return config, nil
}
