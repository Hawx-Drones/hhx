package main

import (
	"fmt"
	"hhx/internal/commands"
	"hhx/internal/config"
	"os"
	"path/filepath"
)

func main() {
	// Create config directory if it doesn't exist
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".hhx")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load(filepath.Join(configDir, "config.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		// We'll continue without config and create it when needed
	}

	// Execute root command
	if err := commands.Execute(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
