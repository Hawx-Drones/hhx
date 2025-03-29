package commands

import (
	"fmt"
	"hhx/internal/config"
	"hhx/internal/models"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var unstageCmd = &cobra.Command{
	Use:   "unstage [file/directory]",
	Short: "Unstage files",
	Long:  `Remove files from the staging area.`,
	Example: `  hhx unstage file.txt      # Unstage a single file
  hhx unstage directory/   # Unstage all files in a directory
  hhx unstage .            # Unstage all files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Println("Error: no files specified")
			err := cmd.Usage()
			if err != nil {
				fmt.Println("error displaying usage:", err)
			}
			return nil
		}

		// Find repository root
		repoRoot, err := findRepoRoot()
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

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			fmt.Println("error loading index:", err)
			return nil
		}

		// Process each argument
		for _, arg := range args {
			// Get absolute path
			path := arg
			if !filepath.IsAbs(path) {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Println("error getting current directory:", err)
					return nil
				}
				path = filepath.Join(cwd, path)
			}

			// Check if the path exists
			info, err := os.Stat(path)
			if os.IsNotExist(err) {
				// Try to find the file in the index by relative path
				relPath, err := filepath.Rel(repoRoot, path)
				if err != nil {
					fmt.Println("error getting relative path for", arg, ":", err)
					return nil
				}
				relPath = filepath.ToSlash(relPath)

				// Unstage the file
				index.UnstageFile(path)
				fmt.Printf("Unstaged %s\n", relPath)
				continue
			} else if err != nil {
				fmt.Println("error accessing", arg, ":", err)
				return nil
			}

			if info.IsDir() {
				// Unstage all files in the directory
				filepath.Walk(path, func(walkPath string, walkInfo os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					if !walkInfo.IsDir() {
						index.UnstageFile(walkPath)
						relPath, _ := filepath.Rel(repoRoot, walkPath)
						relPath = filepath.ToSlash(relPath)
						fmt.Printf("Unstaged %s\n", relPath)
					}

					return nil
				})
			} else {
				// Unstage a single file
				index.UnstageFile(path)
				relPath, _ := filepath.Rel(repoRoot, path)
				relPath = filepath.ToSlash(relPath)
				fmt.Printf("Unstaged %s\n", relPath)
			}
		}

		// Save the index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			fmt.Println("error saving index:", err)
			return nil
		}

		fmt.Println("Files unstaged successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unstageCmd)
}
