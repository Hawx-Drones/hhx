package commands

import (
	"fmt"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"

	"github.com/spf13/cobra"
)

var projectLinkCmd = &cobra.Command{
	Use:   "link [project_name]",
	Short: "Link repository to a project",
	Long:  "Link the current repository to a project on the remote server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			fmt.Println("Error: not in a hhx repository. Initialize one first with 'hhx init'")
			return nil
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			fmt.Println("Error loading repository config:", err)
			return nil
		}

		// Load global config and authenticate
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

		tokenStore := models.NewTokenStore(globalConfigDir)
		token, err := tokenStore.GetToken()
		if err != nil || token == "" {
			fmt.Println("You are not logged in. Please log in first with 'hhx account login'.")
			return nil
		}

		// Verify the project exists using the name-based lookup
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		project, err := client.GetProjectByName(projectName)
		if err != nil {
			fmt.Printf("Error: couldn't find project '%s'. Please check the name or create it first.\n", projectName)
			return nil
		}

		repoConfig.ProjectName = project.Name
		repoConfig.ProjectID = project.ID

		if err := config.SaveRepoConfig(repoConfig); err != nil {
			fmt.Println("Error saving repository config:", err)
			return nil
		}

		fmt.Printf("Repository linked to project '%s' (ID: %s)\n", project.Name, project.ID)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectLinkCmd)
}
