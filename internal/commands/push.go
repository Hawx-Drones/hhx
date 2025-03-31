package commands

import (
	"fmt"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"hhx/internal/util"
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
  hhx push --project=proj-name        # Push to a specific project (overrides the linked project)
  hhx push all                        # Push all files to default collection on default remote
  hhx push origin all                 # Push all files to default collection on specified remote
  hhx push --collection=my-models all # Push all files to specific collection on default remote`,
	RunE: func(cmd *cobra.Command, args []string) error {
		remote := ""
		pushAll := false

		for _, arg := range args {
			if arg == "all" {
				pushAll = true
			} else if remote == "" {
				remote = arg
			} else {
				fmt.Println("Error: unexpected argument:", arg)
				err := cmd.Usage()
				if err != nil {
					fmt.Println("error displaying usage:", err)
				}
				return nil
			}
		}

		collectionName, _ := cmd.Flags().GetString("collection")
		projectName, _ := cmd.Flags().GetString("project")

		repoRoot, err := findRepoRoot()
		if err != nil {
			fmt.Println("could not find repo root:", err)
			return nil
		}

		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			fmt.Println("error loading repository config:", err)
			return nil
		}

		if remote == "" {
			remote = repoConfig.CurrentRemote
		}

		remoteURL, ok := repoConfig.Remotes[remote]
		if !ok {
			fmt.Println("Error: unknown remote:", remote)
			return nil
		}

		// Determine which project to use
		activeProject := repoConfig.ProjectName
		if projectName != "" {
			activeProject = projectName
		}

		if activeProject == "" {
			fmt.Println("Error: no project specified or linked. Use --project to specify a project or link a project with 'hhx project link'")
			return nil
		}

		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			fmt.Println("error loading index:", err)
			return nil
		}

		var collection *models.Collection
		if collectionName != "" {
			collection, err = index.GetCollection(collectionName)
			if err != nil {
				fmt.Println("Error: collection not found:", collectionName)
				return nil
			}
		} else {
			collection, err = index.GetDefaultCollection()
			if err != nil {
				fmt.Println("Error: no default collection set. Use --collection to specify or set a default with 'hhx collection set-default'")
				return nil
			}
		}

		var filesToPush []*models.File

		if pushAll {
			// If pushing all, stage all unstaged files first
			newFiles, modifiedFiles, _, err := index.ScanWorkingDirectory()
			if err != nil {
				fmt.Println("error scanning working directory:", err)
				return nil
			}

			// Stage all new and modified files
			for _, file := range append(newFiles, modifiedFiles...) {
				if err := index.StageFile(file.FullPath(repoRoot)); err != nil {
					fmt.Println("error staging file", file.Path, ":", err)
					return nil
				}
			}
		}

		// Get staged files
		filesToPush = index.GetStagedFiles()
		if len(filesToPush) == 0 {
			fmt.Println("No files to push.")
			return nil
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("error getting home directory:", err)
			return nil
		}
		configDir := filepath.Join(homeDir, ".hhx")
		tokenStore := models.NewTokenStore(configDir)
		client := api.NewClient(remoteURL, tokenStore)
		if client.AuthToken == "" {
			fmt.Println("Error: not logged in. Please run 'hhx login' first")
			return nil
		}

		// Push files to the specific project and collection
		fmt.Printf("Pushing %d files to project '%s', collection '%s' on '%s'...\n",
			len(filesToPush), activeProject, collection.Name, remote)
		startTime := time.Now()

		// Add project and collection information to the push request
		resp, err := client.PushFilesToProjectCollection(repoRoot, filesToPush, activeProject, collection)
		if err != nil {
			fmt.Println("push failed:", err)
			return nil
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

		if err := index.Save(repoConfig.IndexPath); err != nil {
			fmt.Println("error saving index:", err)
			return nil
		}

		// Print summary
		duration := time.Since(startTime).Round(time.Millisecond)
		fileSize := int64(0)
		for _, file := range resp.UploadedFiles {
			fileSize += file.Size
		}

		fmt.Printf("\nUploaded %d files (%s) to project '%s', collection '%s' in %s\n",
			len(resp.UploadedFiles),
			util.FormatSize(fileSize),
			activeProject,
			collection.Name,
			duration,
		)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().Bool("non-interactive", false, "Do not prompt for login")
	pushCmd.Flags().String("collection", "", "Collection to push to (defaults to the default collection)")
	pushCmd.Flags().String("project", "", "Project to push to (overrides the linked project)")
}
