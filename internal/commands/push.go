package commands

import (
	"fmt"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [remote] [all]",
	Short: "Upload files to the remote server",
	Long:  `Upload staged files to the remote server.`,
	Example: `  hhx push                            # Push staged files to default collection on default remote
  hhx push origin                     # Push staged files to default collection on specified remote
  hhx push --collection=my-models     # Push staged files to specific collection on default remote
  hhx push all                        # Push all files to default collection on default remote
  hhx push origin all                 # Push all files to default collection on specified remote
  hhx push --collection=my-models all # Push all files to specific collection on default remote`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse arguments
		remote := ""
		pushAll := false

		for _, arg := range args {
			if arg == "all" {
				pushAll = true
			} else if remote == "" {
				remote = arg
			} else {
				return fmt.Errorf("unexpected argument: %s", arg)
			}
		}

		// Get collection flag
		collectionName, _ := cmd.Flags().GetString("collection")

		// Find repository root
		repoRoot, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Determine remote to use
		if remote == "" {
			remote = repoConfig.CurrentRemote
		}

		remoteURL, ok := repoConfig.Remotes[remote]
		if !ok {
			return fmt.Errorf("unknown remote: %s", remote)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Determine collection to use
		var collection *models.Collection
		if collectionName != "" {
			// Use specified collection
			collection, err = index.GetCollection(collectionName)
			if err != nil {
				return fmt.Errorf("collection not found: %s", collectionName)
			}
		} else {
			// Use default collection
			collection, err = index.GetDefaultCollection()
			if err != nil {
				return fmt.Errorf("no default collection set. Use --collection to specify or set a default with 'hhx collection set-default'")
			}
		}

		// Get files to push
		var filesToPush []*models.File

		if pushAll {
			// If pushing all, stage all unstaged files first
			newFiles, modifiedFiles, _, err := index.ScanWorkingDirectory()
			if err != nil {
				return fmt.Errorf("error scanning working directory: %w", err)
			}

			// Stage all new and modified files
			for _, file := range append(newFiles, modifiedFiles...) {
				if err := index.StageFile(file.FullPath(repoRoot)); err != nil {
					return fmt.Errorf("error staging file %s: %w", file.Path, err)
				}
			}
		}

		// Get staged files
		filesToPush = index.GetStagedFiles()
		if len(filesToPush) == 0 {
			fmt.Println("No files to push.")
			return nil
		}

		// Ensure we have a valid auth token
		if globalConfig.AuthToken == "" {
			// Check if we're in interactive mode
			if !cmd.Flags().Changed("non-interactive") {
				// Prompt for login
				var email, password string
				fmt.Print("Email: ")
				fmt.Scanln(&email)
				fmt.Print("Password: ")
				// In a real implementation, you would use a library to hide the password input
				fmt.Scanln(&password)

				// Login
				client := api.NewClient(remoteURL, "")
				auth, err := client.Login(email, password)
				if err != nil {
					return fmt.Errorf("login failed: %w", err)
				}

				// Save auth token
				globalConfig.AuthToken = auth.Token
				globalConfig.UserID = auth.UserID
				globalConfig.Email = auth.Email

				// Save config
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("error getting home directory: %w", err)
				}
				configPath := filepath.Join(homeDir, ".hhx", "config.json")
				if err := globalConfig.Save(configPath); err != nil {
					return fmt.Errorf("error saving config: %w", err)
				}
			} else {
				return fmt.Errorf("not logged in (use --non-interactive=false to login)")
			}
		}

		// Create API client
		client := api.NewClient(remoteURL, globalConfig.AuthToken)

		// Push files
		fmt.Printf("Pushing %d files to collection '%s' on '%s'...\n", len(filesToPush), collection.Name, remote)
		startTime := time.Now()

		// Add collection information to the push request
		resp, err := client.PushFilesToCollection(repoRoot, filesToPush, collection)
		if err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		// Process response
		if len(resp.Errors) > 0 {
			fmt.Println("Some files failed to upload:")
			for _, uploadErr := range resp.Errors {
				color.Red("  %s: %s\n", uploadErr.Path, uploadErr.Error)
			}
		}

		// Update index with new remote URLs
		for _, uploaded := range resp.UploadedFiles {
			index.MarkSynced(uploaded.Path, uploaded.RemoteURL)
		}

		// Save index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			return fmt.Errorf("error saving index: %w", err)
		}

		// Print summary
		duration := time.Since(startTime).Round(time.Millisecond)
		fileSize := int64(0)
		for _, file := range resp.UploadedFiles {
			fileSize += file.Size
		}

		fmt.Printf("\nUploaded %d files (%s) to collection '%s' in %s\n",
			len(resp.UploadedFiles),
			formatSize(fileSize),
			collection.Name,
			duration,
		)

		return nil
	},
}

// formatSize converts bytes into a human-readable string with appropriate unit suffix
func formatSize(bytes int64) string {
	const (
		_        = iota // ignore first value by assigning to blank identifier
		KB int64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= TB:
		unit = "TB"
		value = float64(bytes) / float64(TB)
	case bytes >= GB:
		unit = "GB"
		value = float64(bytes) / float64(GB)
	case bytes >= MB:
		unit = "MB"
		value = float64(bytes) / float64(MB)
	case bytes >= KB:
		unit = "KB"
		value = float64(bytes) / float64(KB)
	default:
		unit = "bytes"
	}

	if unit == "bytes" {
		return fmt.Sprintf("%d %s", bytes, unit)
	}
	return fmt.Sprintf("%.2f %s", value, unit)
}

func init() {
	pushCmd.Flags().Bool("non-interactive", false, "Do not prompt for login")
	pushCmd.Flags().String("collection", "", "Collection to push to (defaults to the default collection)")
}
