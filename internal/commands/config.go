package commands

import (
	"fmt"
	"hhx/internal/config"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configPathsCmd = &cobra.Command{
	Use:   "config-paths",
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
	rootCmd.AddCommand(configPathsCmd)
}
