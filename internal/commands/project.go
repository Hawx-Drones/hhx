package commands

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"time"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  "Create, list, update, and delete projects",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long:  "Create a new project for organizing buckets",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration and check authentication
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
			fmt.Println("You are not logged in. Please log in first.")
			return nil
		}

		// Get project details from flags or prompt
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		// If name wasn't provided via flag, prompt for it
		if name == "" {
			fmt.Print("Project name: ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				name = scanner.Text()
			}

			if name == "" {
				fmt.Println("Error: Project name is required")
				return nil
			}
		}

		// If description wasn't provided via flag, prompt for it
		if description == "" {
			fmt.Print("Project description (optional): ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				description = scanner.Text()
			}
		}

		// Create the project
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		project, err := client.CreateProject(name, description)
		if err != nil {
			fmt.Println("Error creating project:", err)
			return nil
		}

		fmt.Printf("Project created successfully!\n")
		fmt.Printf("ID: %s\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		fmt.Printf("Description: %s\n", project.Description)

		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  "List all projects for the current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration and check authentication
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
			fmt.Println("You are not logged in. Please log in first.")
			return nil
		}

		// List projects
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		projects, err := client.ListProjects()
		if err != nil {
			fmt.Println("Error listing projects:", err)
			return nil
		}

		if len(projects) == 0 {
			fmt.Println("No projects found. Create one with 'hhx project create'")
			return nil
		}

		fmt.Printf("Projects:\n\n")
		for i, project := range projects {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, project.Name, project.ID)
			fmt.Printf("   Description: %s\n", project.Description)
			fmt.Printf("   Created: %s\n", project.CreatedAt.Format(time.RFC1123))
			fmt.Printf("   Updated: %s\n", project.UpdatedAt.Format(time.RFC1123))
			fmt.Println()
		}

		return nil
	},
}

var projectShowCmd = &cobra.Command{
	Use:   "show [project_id]",
	Short: "Show project details",
	Long:  "Show detailed information about a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Load configuration and check authentication
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
			fmt.Println("You are not logged in. Please log in first.")
			return nil
		}

		// Get project details
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		project, err := client.GetProject(projectID)
		if err != nil {
			fmt.Println("Error getting project:", err)
			return nil
		}

		fmt.Printf("Project Details:\n\n")
		fmt.Printf("ID: %s\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)
		fmt.Printf("Description: %s\n", project.Description)
		fmt.Printf("Created: %s\n", project.CreatedAt.Format(time.RFC1123))
		fmt.Printf("Updated: %s\n", project.UpdatedAt.Format(time.RFC1123))

		return nil
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project_id]",
	Short: "Update project",
	Long:  "Update a project's details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Load configuration and check authentication
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
			fmt.Println("You are not logged in. Please log in first.")
			return nil
		}

		// Get current project details
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		project, err := client.GetProject(projectID)
		if err != nil {
			fmt.Println("Error getting project:", err)
			return nil
		}

		// Get updated details
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		// If neither name nor description provided, prompt for updating
		if name == "" && description == "" {
			// Prompt for new name
			fmt.Printf("Name [%s]: ", project.Name)
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				newName := scanner.Text()
				if newName != "" {
					name = newName
				}
			}

			// Prompt for new description
			fmt.Printf("Description [%s]: ", project.Description)
			if scanner.Scan() {
				newDesc := scanner.Text()
				if newDesc != "" {
					description = newDesc
				}
			}
		}

		// Only update if something has changed
		if name == "" && description == "" {
			fmt.Println("No changes to make. Project remains unchanged.")
			return nil
		}

		// Use existing values if not updating
		if name == "" {
			name = project.Name
		}
		if description == "" {
			description = project.Description
		}

		// Update the project
		updatedProject, err := client.UpdateProject(projectID, name, description)
		if err != nil {
			fmt.Println("Error updating project:", err)
			return nil
		}

		fmt.Println("Project updated successfully!")
		fmt.Printf("ID: %s\n", updatedProject.ID)
		fmt.Printf("Name: %s\n", updatedProject.Name)
		fmt.Printf("Description: %s\n", updatedProject.Description)

		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project_id]",
	Short: "Delete project",
	Long:  "Delete a project and optionally its associated buckets",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Load configuration and check authentication
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
			fmt.Println("You are not logged in. Please log in first.")
			return nil
		}

		// Get current project details
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		project, err := client.GetProject(projectID)
		if err != nil {
			fmt.Println("Error getting project:", err)
			return nil
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Are you sure you want to delete project '%s'? (y/n): ", project.Name)
			var confirmation string
			_, err = fmt.Scanln(&confirmation)
			if err != nil || (confirmation != "y" && confirmation != "Y") {
				fmt.Println("Project deletion cancelled.")
				return nil
			}
		}

		// Delete the project
		err = client.DeleteProject(projectID)
		if err != nil {
			fmt.Println("Error deleting project:", err)
			return nil
		}

		fmt.Println("Project deleted successfully!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectShowCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)

	projectCreateCmd.Flags().String("name", "", "Project name")
	projectCreateCmd.Flags().String("description", "", "Project description")

	projectUpdateCmd.Flags().String("name", "", "Project name")
	projectUpdateCmd.Flags().String("description", "", "Project description")

	projectDeleteCmd.Flags().Bool("force", false, "Force deletion without confirmation")
}
