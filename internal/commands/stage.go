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
	Example: `  hhx stage file.txt       # Stage a single file
  hhx stage directory/     # Stage all files in a directory
  hhx stage .              # Stage all files in the current directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("Error: no files specified")
			err := cmd.Usage()
			if err != nil {
				fmt.Println("error displaying usage: %w", err)
				return err
			}
			return nil
		}

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			fmt.Println("Could not find repo root: %w", err)
			return nil
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			fmt.Println("error loading repository config: %w", err)
			return nil
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			fmt.Println("error loading index: %w", err)
			return nil
		}

		// Process each argument
		for _, arg := range args {
			// Get absolute path
			path := arg
			if !filepath.IsAbs(path) {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Println("error getting current directory: %w", err)
					return nil
				}
				path = filepath.Join(cwd, path)
			}

			// Check if the path exists
			info, err := os.Stat(path)
			if err != nil {
				fmt.Println("error accessing %s: %w", arg, err)
				return nil
			}

			if info.IsDir() {
				// Stage all files in the directory
				fmt.Printf("Staging files in directory %s...\n", arg)
				if err := index.StageDirectory(path); err != nil {
					fmt.Println("error staging directory %s: %w", arg, err)
					return nil
				}
			} else {
				// Stage a single file
				fmt.Printf("Staging file %s...\n", arg)
				if err := index.StageFile(path); err != nil {
					fmt.Println("error staging file %s: %w", arg, err)
					return nil
				}
			}
		}

		// Save the index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			fmt.Println("error saving index: %w", err)
			return nil
		}

		fmt.Println("Files staged successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stageCmd)
}
