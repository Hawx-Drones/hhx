package commands

import (
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var stageCmd = &cobra.Command{
	Use:   "stage [file/directory]",
	Short: "Stage files for upload",
	Long:  `Stage files or directories for upload to the remote server.`,
	Example: `  hhx stage file.txt        # Stage a single file
  hhx stage directory/     # Stage all files in a directory
  hhx stage .              # Stage all files in the current directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("no files specified")
		}

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			return err
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			return fmt.Errorf("error loading repository config: %w", err)
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			return fmt.Errorf("error loading index: %w", err)
		}

		// Process each argument
		for _, arg := range args {
			// Get absolute path
			path := arg
			if !filepath.IsAbs(path) {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("error getting current directory: %w", err)
				}
				path = filepath.Join(cwd, path)
			}

			// Check if the path exists
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("error accessing %s: %w", arg, err)
			}

			if info.IsDir() {
				// Stage all files in the directory
				fmt.Printf("Staging files in directory %s...\n", arg)
				if err := index.StageDirectory(path); err != nil {
					return fmt.Errorf("error staging directory %s: %w", arg, err)
				}
			} else {
				// Stage a single file
				fmt.Printf("Staging file %s...\n", arg)
				if err := index.StageFile(path); err != nil {
					return fmt.Errorf("error staging file %s: %w", arg, err)
				}
			}
		}

		// Save the index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			return fmt.Errorf("error saving index: %w", err)
		}

		fmt.Println("Files staged successfully.")
		return nil
	},
}

func init() {
	// Add flags if necessary
}
