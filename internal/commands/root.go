package commands

import (
	"fmt"
	"hhx/internal/config"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var globalConfig *config.Config

var rootCmd = &cobra.Command{
	Use:   "hhx",
	Short: "Headless Hawx - A tool for managing database and storage resources",
	Long: `Headless Hawx (hhx) is a command-line tool for managing database tables and storage buckets.
It provides a git-like interface for tracking changes, staging files, and pushing them to remote servers.`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute(cfg *config.Config) error {
	globalConfig = cfg
	return rootCmd.Execute()
}

// findRepoRoot finds the root directory of the repository
func findRepoRoot() (string, error) {
	// Start from current directory and traverse up until we find .hhx directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		hhxDir := filepath.Join(dir, ".hhx")
		if _, err := os.Stat(hhxDir); err == nil {
			return dir, nil
		}

		// Stop if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("not in a hhx repository (or any parent directory)")
}

func init() {
	// Add all commands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stageCmd)
	rootCmd.AddCommand(unstageCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(collectionCmd)
}
