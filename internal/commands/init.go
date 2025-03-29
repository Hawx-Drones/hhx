package commands

import (
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new repository",
	Long:  `Initialize a new hhx repository in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("error getting current directory:", err)
			return nil
		}

		// Check if .hhx directory already exists
		hhxDir := filepath.Join(cwd, ".hhx")
		if _, err := os.Stat(hhxDir); err == nil {
			fmt.Println("Error: repository already initialized")
			return nil
		}

		// Create .hhx directory
		if err := os.MkdirAll(hhxDir, 0755); err != nil {
			fmt.Println("error creating .hhx directory:", err)
			return nil
		}

		// Create index file
		indexPath := filepath.Join(hhxDir, "index.json")
		index := models.NewIndex(cwd)

		// Create a default collection
		defaultCollection := &models.Collection{
			Name: "default",
			Type: models.CollectionTypeBucket,
			Path: "default",
		}

		// Add the default collection to the index
		if err := index.AddCollection(defaultCollection); err != nil {
			fmt.Println("error creating default collection:", err)
			return nil
		}

		// Save the index
		if err := index.Save(indexPath); err != nil {
			fmt.Println("error creating index file:", err)
			return nil
		}

		// Create repository config
		repoConfig := &config.RepoConfig{
			Remotes: map[string]string{
				"origin": globalConfig.ServerURL,
			},
			CurrentRemote: "origin",
			IndexPath:     indexPath,
		}

		// Save repository config
		repoConfigPath := filepath.Join(hhxDir, "config.json")
		fmt.Printf("repoConfigPath: %v\n", repoConfigPath)
		if err := config.SaveRepoConfig(repoConfig); err != nil {
			fmt.Println("error creating repository config:", err)
			return nil
		}

		fmt.Println("Initialized empty hhx repository in", hhxDir)
		fmt.Println("Created default collection for storing files")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
