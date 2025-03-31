package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"hhx/internal/util"
	"strings"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage buckets",
	Long:  "Create, list, update, and delete storage buckets within a specific project.",
}

// storageListCmd
var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all storage buckets in a project",
	Long:  `List all storage buckets for the specified project (or the project linked to this repo).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		buckets, err := client.ListBuckets(projectID)
		if err != nil {
			fmt.Println("Error listing buckets:", err)
			return nil
		}

		if len(buckets) == 0 {
			fmt.Println("No buckets found.")
			return nil
		}

		fmt.Printf("Buckets for project %s:\n", projectID)
		for _, bucket := range buckets {
			fmt.Printf("  - %s (Public: %t)\n", bucket.Name, bucket.Public)
		}
		return nil
	},
}

// storageCreateCmd
var storageCreateCmd = &cobra.Command{
	Use:   "create [bucketName]",
	Short: "Create a new storage bucket in a project",
	Long:  `Create a new storage bucket under the specified project (or the project linked to this repo).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

		public, _ := cmd.Flags().GetBool("public")
		fileSizeLimit, _ := cmd.Flags().GetInt64("file-size-limit")
		allowedFileTypes, _ := cmd.Flags().GetString("allowed-file-types")

		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		// 3) Create client and call new CreateBucket(projectID, ...)
		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		bucket, err := client.CreateBucket(projectID, bucketName, public, fileSizeLimit, allowedFileTypes)
		if err != nil {
			fmt.Println("Error creating bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' created successfully in project '%s'\n", bucket.Name, projectID)
		return nil
	},
}

// storageGetCmd
var storageGetCmd = &cobra.Command{
	Use:   "get [bucketName]",
	Short: "Get details of a bucket in a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		bucket, err := client.GetBucket(projectID, bucketName)
		if err != nil {
			fmt.Println("Error getting bucket:", err)
			return nil
		}

		fmt.Printf("Bucket Details for '%s' in project '%s':\n", bucketName, projectID)
		fmt.Printf("  ID: %s\n", bucket.ID)
		fmt.Printf("  Name: %s\n", bucket.Name)
		fmt.Printf("  Public: %t\n", bucket.Public)
		fmt.Printf("  CreatorID: %s\n", bucket.CreatorID)
		fmt.Printf("  Created At: %s\n", bucket.CreatedAt)
		fmt.Printf("  Updated At: %s\n", bucket.UpdatedAt)
		fmt.Printf("  Allowed Types: %v\n", bucket.AllowedFileTypes)
		return nil
	},
}

// storageUpdateCmd
var storageUpdateCmd = &cobra.Command{
	Use:   "update [bucketName]",
	Short: "Update settings for a bucket in a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

		// If neither public nor file-size-limit nor allowed-file-types is changed, exit
		hasUpdate := cmd.Flags().Changed("public") ||
			cmd.Flags().Changed("file-size-limit") ||
			cmd.Flags().Changed("allowed-file-types")
		if !hasUpdate {
			fmt.Println("Error: at least one flag must be specified to update.")
			return nil
		}

		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		// Build the patch
		public, _ := cmd.Flags().GetBool("public")
		fileSizeLimit, _ := cmd.Flags().GetInt64("file-size-limit")
		allowedFileTypes, _ := cmd.Flags().GetString("allowed-file-types")

		updates := make(map[string]interface{})
		if cmd.Flags().Changed("public") {
			updates["public"] = public
		}
		if cmd.Flags().Changed("file-size-limit") {
			updates["fileSizeLimit"] = fileSizeLimit
		}
		if cmd.Flags().Changed("allowed-file-types") {
			updates["allowedFileTypes"] = strings.Split(allowedFileTypes, ",")
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		err = client.UpdateBucket(projectID, bucketName, updates)
		if err != nil {
			fmt.Println("Error updating bucket:", err)
			return nil
		}
		fmt.Printf("Bucket '%s' updated successfully in project '%s'\n", bucketName, projectID)
		return nil
	},
}

// storageDeleteCmd
var storageDeleteCmd = &cobra.Command{
	Use:   "delete [bucketName]",
	Short: "Delete a bucket from a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete bucket '%s'? This action cannot be undone. [y/N]: ", bucketName)
			var confirm string
			scanln, err := fmt.Scanln(&confirm)
			if err != nil {
				fmt.Println("Error reading input:", scanln)
				return nil
			}
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		err = client.DeleteBucket(projectID, bucketName)
		if err != nil {
			fmt.Println("Error deleting bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' deleted successfully from project '%s'\n", bucketName, projectID)
		return nil
	},
}

// storageEmptyCmd
var storageEmptyCmd = &cobra.Command{
	Use:   "empty [bucketName]",
	Short: "Remove all files from a bucket (but keep the bucket)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to empty bucket '%s'? This action cannot be undone. [y/N]: ", bucketName)
			var confirm string
			_, err := fmt.Scanln(&confirm)
			if err != nil {
				fmt.Println()
				return err
			}
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}

		projectID, err := resolveProjectID(cmd)
		if err != nil {
			fmt.Println("Error determining project:", err)
			return nil
		}
		if projectID == "" {
			fmt.Println("Error: No project specified or linked.")
			return nil
		}

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
			fmt.Println("You are not logged in")
			return nil
		}

		client := api.NewClient(globalConfig.ServerURL, tokenStore)
		err = client.EmptyBucket(projectID, bucketName)
		if err != nil {
			fmt.Println("Error emptying bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' emptied successfully in project '%s'\n", bucketName, projectID)
		return nil
	},
}

// init
func init() {
	rootCmd.AddCommand(storageCmd)

	// Register subcommands
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageGetCmd)
	storageCmd.AddCommand(storageUpdateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageEmptyCmd)

	// Common project flag on storageCmd and subcommands
	for _, sc := range storageCmd.Commands() {
		sc.Flags().String("project", "", "Project name or ID (overrides linked project)")
	}

	// Additional flags
	storageCreateCmd.Flags().Bool("public", false, "Whether the bucket should be publicly accessible")
	storageCreateCmd.Flags().Int64("file-size-limit", 0, "Maximum file size in bytes (0 for no limit)")
	storageCreateCmd.Flags().String("allowed-file-types", "", "Comma-separated list of allowed file extensions")

	storageUpdateCmd.Flags().Bool("public", false, "Whether the bucket should be publicly accessible")
	storageUpdateCmd.Flags().Int64("file-size-limit", 0, "Maximum file size in bytes (0 for no limit)")
	storageUpdateCmd.Flags().String("allowed-file-types", "", "Comma-separated list of allowed file extensions")

	storageDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	storageEmptyCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}

// ResolveProjectID resolves the project ID from command line flags or local config
func resolveProjectID(cmd *cobra.Command) (string, error) {
	// 1) Check if user passed --project explicitly
	projectFlag, _ := cmd.Flags().GetString("project")
	if projectFlag != "" {
		// Could be a direct ID or a project name
		// If it looks like a UUID, we assume it’s the ID
		// else we do a lookup
		if util.IsUUID(projectFlag) {
			return projectFlag, nil
		}
		// If it's not a UUID, we look it up
		return lookupProjectIDByName(projectFlag)
	}

	// 2) If no flag, check local repo config
	repoConfig, err := config.LoadRepoConfig()
	if err != nil {
		return "", err
	}
	if repoConfig.ProjectID != "" {
		return repoConfig.ProjectID, nil
	}
	if repoConfig.ProjectName != "" {
		// We only have the name, so lookup
		return lookupProjectIDByName(repoConfig.ProjectName)
	}

	// If nothing found, return blank
	return "", nil
}

// lookupProjectIDByName calls the API to find a project by name
func lookupProjectIDByName(name string) (string, error) {
	globalConfigDir, err := config.GetGlobalConfigDir()
	if err != nil {
		return "", err
	}
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return "", err
	}
	tokenStore := models.NewTokenStore(globalConfigDir)
	token, err := tokenStore.GetToken()
	if err != nil || token == "" {
		return "", fmt.Errorf("not logged in")
	}
	client := api.NewClient(globalConfig.ServerURL, tokenStore)
	proj, err := client.GetProjectByName(name)
	if err != nil {
		return "", err
	}
	return proj.ID, nil
}
