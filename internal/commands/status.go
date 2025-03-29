package commands

import (
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Long:  `Display the state of the working directory and the staging area.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			fmt.Println("could not find repo root:", err)
			return nil
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			fmt.Println("error loading repository config:", err)
			return nil
		}

		if repoConfig.CurrentRemote == "" {
			fmt.Println("Error: no remote repository configured")
			return nil
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			fmt.Println("error loading index:", err)
			return nil
		}

		// Scan working directory for changes
		newFiles, modifiedFiles, deletedFiles, err := index.ScanWorkingDirectory()
		if err != nil {
			fmt.Println("error scanning working directory:", err)
			return nil
		}

		// Get staged files
		stagedFiles := index.GetStagedFiles()

		// Format output
		fmt.Printf("On remote: %s (%s)\n", repoConfig.CurrentRemote, repoConfig.Remotes[repoConfig.CurrentRemote])
		fmt.Println()

		// Changes to be uploaded
		if len(stagedFiles) > 0 {
			fmt.Println("Changes to be uploaded:")
			fmt.Println("  (use \"hhx unstage <file>...\" to unstage)")
			fmt.Println()

			// Sort files by path
			sort.Slice(stagedFiles, func(i, j int) bool {
				return stagedFiles[i].Path < stagedFiles[j].Path
			})

			// Print new files
			for _, file := range stagedFiles {
				if file.RemoteURL == "" {
					color.Green("\tnew file:   %s\n", file.Path)
				} else {
					color.Yellow("\tmodified:   %s\n", file.Path)
				}
			}
			fmt.Println()
		}

		// Changes not staged for upload
		if len(modifiedFiles) > 0 || len(deletedFiles) > 0 {
			fmt.Println("Changes not staged for upload:")
			fmt.Println("  (use \"hhx stage <file>...\" to update what will be uploaded)")
			fmt.Println()

			// Sort files by path
			sort.Slice(modifiedFiles, func(i, j int) bool {
				return modifiedFiles[i].Path < modifiedFiles[j].Path
			})
			sort.Slice(deletedFiles, func(i, j int) bool {
				return deletedFiles[i].Path < deletedFiles[j].Path
			})

			// Print modified files
			for _, file := range modifiedFiles {
				color.Yellow("\tmodified:   %s\n", file.Path)
			}

			// Print deleted files
			for _, file := range deletedFiles {
				color.Red("\tdeleted:    %s\n", file.Path)
			}
			fmt.Println()
		}

		// Untracked files
		if len(newFiles) > 0 {
			fmt.Println("Untracked files:")
			fmt.Println("  (use \"hhx stage <file>...\" to include in what will be uploaded)")
			fmt.Println()

			// Sort files by path
			sort.Slice(newFiles, func(i, j int) bool {
				return newFiles[i].Path < newFiles[j].Path
			})

			// Group files by directory for better readability
			filesByDirectory := make(map[string][]string)

			for _, file := range newFiles {
				dir := "."
				if idx := strings.LastIndex(file.Path, "/"); idx != -1 {
					dir = file.Path[:idx]
				}
				filesByDirectory[dir] = append(filesByDirectory[dir], file.Path)
			}

			// Print untracked files
			dirs := make([]string, 0, len(filesByDirectory))
			for dir := range filesByDirectory {
				dirs = append(dirs, dir)
			}
			sort.Strings(dirs)

			for _, dir := range dirs {
				files := filesByDirectory[dir]
				sort.Strings(files)
				for _, file := range files {
					color.Red("\t%s\n", file)
				}
			}
			fmt.Println()
		}

		// Summary
		stagedCount := len(stagedFiles)
		notStagedCount := len(modifiedFiles) + len(deletedFiles)
		untrackedCount := len(newFiles)

		if stagedCount == 0 && notStagedCount == 0 && untrackedCount == 0 {
			fmt.Println("No changes (working directory clean)")
		} else {
			parts := []string{}
			if stagedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d to be uploaded", stagedCount))
			}
			if notStagedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d not staged", notStagedCount))
			}
			if untrackedCount > 0 {
				parts = append(parts, fmt.Sprintf("%d untracked", untrackedCount))
			}
			fmt.Printf("%s\n", strings.Join(parts, ", "))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
