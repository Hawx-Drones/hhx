package commands

import (
	"fmt"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project]",
	Short: "Initialize a new repository linked to a project",
	Long:  `Initialize a new hhx repository in the current directory and link it to a remote project.`,
	Example: `  hhx init                                          # Initialize a new repository with default settings
  hhx init --project myproject                      # Initialize and link to an existing project
  hhx init --remote customserver                    # Initialize with a specific remote server
  hhx init --project myproject --collection models  # Link to a specific collection`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get current directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current directory:", err)
			return nil
		}

		// Check if .hhx directory already exists
		hhxDir := filepath.Join(cwd, ".hhx")
		if _, err := os.Stat(hhxDir); err == nil {
			fmt.Println("Error: repository already initialized")
			return nil
		}

		// Get flags
		projectName, _ := cmd.Flags().GetString("project")
		remoteName, _ := cmd.Flags().GetString("remote")
		collectionName, _ := cmd.Flags().GetString("collection")

		if remoteName == "" {
			remoteName = "origin"
		}

		// Load global config
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}

		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
			return nil
		}

		// Validate authentication if linking to a project
		if projectName != "" {
			tokenStore := models.NewTokenStore(globalConfigDir)
			token, err := tokenStore.GetToken()
			if err != nil || token == "" {
				fmt.Println("You are not logged in. Please log in first with 'hhx account login'.")
				return nil
			}

			// Verify the project exists on the server
			client := api.NewClient(globalConfig.ServerURL, tokenStore)
			project, err := client.GetProject(projectName)
			if err != nil {
				fmt.Printf("Error: couldn't find project '%s' on the server. Please check the project name or create it first with 'hhx project create'.\n", projectName)
				return nil
			}

			fmt.Printf("Linking to project '%s' (ID: %s)\n", project.Name, project.ID)
		}

		// Create .hhx directory
		if err := os.MkdirAll(hhxDir, 0755); err != nil {
			fmt.Println("Error creating .hhx directory:", err)
			return nil
		}

		// Create index file
		indexPath := filepath.Join(hhxDir, "index.json")
		index := models.NewIndex(cwd)

		// Create a default collection if no specific collection is provided
		defaultCollectionName := "default"
		if collectionName != "" {
			defaultCollectionName = collectionName
		}

		defaultCollection := &models.Collection{
			Name: defaultCollectionName,
			Type: models.CollectionTypeBucket,
			Path: defaultCollectionName,
		}

		// Add the default collection to the index
		if err := index.AddCollection(defaultCollection); err != nil {
			fmt.Println("Error creating default collection:", err)
			return nil
		}

		// Set it as the default collection
		if err := index.SetDefaultCollection(defaultCollectionName); err != nil {
			fmt.Println("Error setting default collection:", err)
			return nil
		}

		// Save the index
		if err := index.Save(indexPath); err != nil {
			fmt.Println("Error creating index file:", err)
			return nil
		}

		// Create repository config
		repoConfig := &config.RepoConfig{
			Remotes: map[string]string{
				remoteName: globalConfig.ServerURL,
			},
			CurrentRemote: remoteName,
			IndexPath:     indexPath,
		}

		// If a project was specified, store it in the config
		if projectName != "" {
			repoConfig.ProjectName = projectName
		}

		// Save repository config
		_ = filepath.Join(hhxDir, "config.json")
		if err := config.SaveRepoConfig(repoConfig); err != nil {
			fmt.Println("Error creating repository config:", err)
			return nil
		}

		fmt.Println("Initialized empty hhx repository in", hhxDir)
		if projectName != "" {
			fmt.Printf("Linked to project '%s' on remote '%s'\n", projectName, remoteName)
		} else {
			fmt.Println("No project linked. You can link a project later with 'hhx project link'.")
		}
		fmt.Printf("Created default collection '%s' for storing files\n", defaultCollectionName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("project", "", "Link to an existing remote project")
	initCmd.Flags().String("remote", "origin", "Name of the remote")
	initCmd.Flags().String("collection", "", "Default collection to use (defaults to 'default')")
}
