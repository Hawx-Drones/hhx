package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage buckets",
	Long:  "Create, list, update, and delete storage buckets",
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all storage buckets",
	Long:  "List all storage buckets in the project",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		buckets, err := client.ListBuckets()
		if err != nil {
			fmt.Println("Error listing buckets:", err)
			return nil
		}

		if len(buckets) == 0 {
			fmt.Println("No buckets found")
			return nil
		}

		fmt.Println("Buckets:")
		for _, bucket := range buckets {
			fmt.Printf("  - %s (Public: %t)\n", bucket.Name, bucket.Public)
		}

		return nil
	},
}

var storageCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new storage bucket",
	Long:  "Create a new storage bucket in the project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		public, _ := cmd.Flags().GetBool("public")
		fileSizeLimit, _ := cmd.Flags().GetInt64("file-size-limit")

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
		bucket, err := client.CreateBucket(bucketName, public, fileSizeLimit)
		if err != nil {
			fmt.Println("Error creating bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' created successfully\n", bucket.Name)
		return nil
	},
}

var storageGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get bucket details",
	Long:  "Get details about a specific bucket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

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
		bucket, err := client.GetBucket(bucketName)
		if err != nil {
			fmt.Println("Error getting bucket:", err)
			return nil
		}

		fmt.Printf("Bucket Details:\n")
		fmt.Printf("  Name: %s\n", bucket.Name)
		fmt.Printf("  ID: %s\n", bucket.ID)
		fmt.Printf("  Public: %t\n", bucket.Public)
		fmt.Printf("  Owner: %s\n", bucket.Owner)
		fmt.Printf("  Created At: %s\n", bucket.CreatedAt)
		fmt.Printf("  Updated At: %s\n", bucket.UpdatedAt)

		return nil
	},
}

var storageUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update bucket settings",
	Long:  "Update settings for a specific bucket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]

		// Only update if flags are explicitly set
		if !cmd.Flags().Changed("public") && !cmd.Flags().Changed("file-size-limit") {
			fmt.Println("Error: at least one flag must be specified")
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

		// Only update the fields that were explicitly set
		public, err := cmd.Flags().GetBool("public")
		var updates struct {
			Public        *bool  `json:"public,omitempty"`
			FileSizeLimit *int64 `json:"fileSizeLimit,omitempty"`
		}

		if err == nil {
			updates.Public = &public
		} else {
			updates.Public = nil
		}

		fileSizeLimit, err := cmd.Flags().GetInt64("file-size-limit")
		if err == nil {
			updates.FileSizeLimit = &fileSizeLimit
		} else {
			updates.FileSizeLimit = nil
		}

		err = client.UpdateBucket(bucketName, updates)
		if err != nil {
			fmt.Println("Error updating bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' updated successfully\n", bucketName)
		return nil
	},
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a bucket",
	Long:  "Delete a storage bucket and all its contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to delete bucket '%s'? This action cannot be undone. [y/N]: ", bucketName)
			var confirm string
			_, err := fmt.Scanln(&confirm)
			if err != nil {
				fmt.Println("Error reading input:", err)
				return nil
			}
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Operation cancelled")
				return nil
			}
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
		err = client.DeleteBucket(bucketName)
		if err != nil {
			fmt.Println("Error deleting bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' deleted successfully\n", bucketName)
		return nil
	},
}

var storageEmptyCmd = &cobra.Command{
	Use:   "empty [name]",
	Short: "Empty a bucket",
	Long:  "Remove all files from a bucket without deleting the bucket itself",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bucketName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to empty bucket '%s'? This action cannot be undone. [y/N]: ", bucketName)
			var confirm string
			_, err := fmt.Scanln(&confirm)
			if err != nil {
				fmt.Println("Error reading input:", err)
				return nil
			}
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Operation cancelled")
				return nil
			}
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
		err = client.EmptyBucket(bucketName)
		if err != nil {
			fmt.Println("Error emptying bucket:", err)
			return nil
		}

		fmt.Printf("Bucket '%s' emptied successfully\n", bucketName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storageCmd)

	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageGetCmd)
	storageCmd.AddCommand(storageUpdateCmd)
	storageCmd.AddCommand(storageDeleteCmd)
	storageCmd.AddCommand(storageEmptyCmd)

	storageCreateCmd.Flags().Bool("public", false, "Whether the bucket should be publicly accessible")
	storageCreateCmd.Flags().Int64("file-size-limit", 0, "Maximum file size in bytes (0 for no limit)")

	storageUpdateCmd.Flags().Bool("public", false, "Whether the bucket should be publicly accessible")
	storageUpdateCmd.Flags().Int64("file-size-limit", 0, "Maximum file size in bytes (0 for no limit)")

	storageDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	storageEmptyCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}
