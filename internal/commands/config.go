package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"hhx/internal/config"
	"os"
	"path/filepath"
)

var (
	// Variables to hold flag values
	serverURL string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage HHX configuration",
	Long:  "View and update HHX configuration settings",
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long:  "Display specific configuration value or all configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// If no argument is provided, show all config
		if len(args) == 0 {
			fmt.Println("Current configuration:")
			fmt.Printf("Server URL: %s\n", cfg.ServerURL)
			if cfg.DefaultRepoPath != "" {
				fmt.Printf("Default Repository Path: %s\n", cfg.DefaultRepoPath)
			}
			if cfg.Email != "" {
				fmt.Printf("Email: %s\n", cfg.Email)
			}
			return nil
		}

		// Show specific config value
		switch args[0] {
		case "server-url":
			fmt.Println(cfg.ServerURL)
		case "default-repo-path":
			fmt.Println(cfg.DefaultRepoPath)
		case "email":
			fmt.Println(cfg.Email)
		default:
			return fmt.Errorf("unknown configuration key: %s", args[0])
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long:  "Update configuration settings like server URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Update configuration based on provided flags
		configUpdated := false

		if serverURL != "" {
			oldURL := cfg.ServerURL
			cfg.ServerURL = serverURL
			fmt.Printf("Server URL updated: %s -> %s\n", oldURL, serverURL)
			configUpdated = true
		}

		// Save configuration if it was updated
		if configUpdated {
			if err := config.SaveGlobalConfig(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
			fmt.Println("Configuration updated successfully.")
		} else {
			fmt.Println("No changes were made to the configuration.")
		}

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Create a new configuration file with default values",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetGlobalConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); err == nil {
			fmt.Println("Configuration file already exists.")
			fmt.Println("Use 'hhx config set' to modify existing configuration.")
			return nil
		}

		// Create default configuration
		cfg := &config.Config{
			ServerURL: "https://api.headlesshawx.io",
		}

		// Override defaults with provided flags
		if serverURL != "" {
			cfg.ServerURL = serverURL
		}

		if err := config.SaveGlobalConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println("Configuration initialized successfully.")
		fmt.Printf("Configuration file created at: %s\n", configPath)
		return nil
	},
}

var configPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Show configuration file paths",
	Long:  "Display paths to global and local configuration files",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			return err
		}
		globalConfigPath := filepath.Join(globalConfigDir, "config.json")
		globalTokenPath := filepath.Join(globalConfigDir, ".auth_token")

		fmt.Println("Global config paths:")
		fmt.Printf("- Config directory: %s\n", globalConfigDir)
		fmt.Printf("- Config file: %s\n", globalConfigPath)
		fmt.Printf("- Auth token file: %s\n", globalTokenPath)

		// Check if current directory has a repo config
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %w", err)
		}

		repoConfigDir := filepath.Join(cwd, ".hhx")
		repoConfigPath := filepath.Join(repoConfigDir, "config.json")

		fmt.Println("\nRepository config paths:")
		fmt.Printf("- Config directory: %s\n", repoConfigDir)
		fmt.Printf("- Config file: %s\n", repoConfigPath)

		// Check existence
		fmt.Println("\nExistence status:")
		if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
			fmt.Println("- Global config file: Does not exist")
		} else {
			fmt.Println("- Global config file: Exists")
		}

		if _, err := os.Stat(globalTokenPath); os.IsNotExist(err) {
			fmt.Println("- Global auth token: Does not exist")
		} else {
			fmt.Println("- Global auth token: Exists")
		}

		if _, err := os.Stat(repoConfigPath); os.IsNotExist(err) {
			fmt.Println("- Repo config file: Does not exist")
		} else {
			fmt.Println("- Repo config file: Exists")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathsCmd)

	configSetCmd.Flags().StringVar(&serverURL, "server-url", "", "Set API server URL")

	configInitCmd.Flags().StringVar(&serverURL, "server-url", "", "Set API server URL")
}
