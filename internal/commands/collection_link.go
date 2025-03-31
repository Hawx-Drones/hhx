package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"hhx/internal/api"
	"hhx/internal/config"
	"hhx/internal/models"
	"io"
	"net/http"
	"time"
)

// collectionLinkCmd represents the collection link command
var collectionLinkCmd = &cobra.Command{
	Use:   "link [name]",
	Short: "Link a collection to a remote storage bucket",
	Long:  `Link a local collection to a remote storage bucket for pushing files.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collectionName := args[0]
		remoteBucketName, _ := cmd.Flags().GetString("bucket")
		createIfMissing, _ := cmd.Flags().GetBool("create")

		// If no bucket name specified, use the collection name
		if remoteBucketName == "" {
			remoteBucketName = collectionName
		}

		// Find repository root
		_, err := findRepoRoot()
		if err != nil {
			fmt.Println("Could not find repo root:", err)
			return nil
		}

		// Load repository config
		repoConfig, err := config.LoadRepoConfig()
		if err != nil {
			fmt.Println("Error loading repository config:", err)
			return nil
		}

		// Load index
		index, err := models.LoadIndex(repoConfig.IndexPath)
		if err != nil {
			fmt.Println("Error loading index:", err)
			return nil
		}

		// Get the collection
		collection, err := index.GetCollection(collectionName)
		if err != nil {
			fmt.Println("Error: collection not found:", collectionName)
			fmt.Println("Create it first with 'hhx collection create'")
			return nil
		}

		// Set up API client
		globalConfigDir, err := config.GetGlobalConfigDir()
		if err != nil {
			fmt.Println("Error getting global config directory:", err)
			return nil
		}
		tokenStore := models.NewTokenStore(globalConfigDir)

		// Load global config
		globalConfig, err := config.LoadGlobalConfig()
		if err != nil {
			fmt.Println("Error loading global config:", err)
			return nil
		}

		// Use current remote URL
		remoteURL := globalConfig.ServerURL
		if repoConfig.CurrentRemote != "" {
			if url, ok := repoConfig.Remotes[repoConfig.CurrentRemote]; ok {
				remoteURL = url
			}
		}

		client := api.NewClient(remoteURL, tokenStore)

		// Check if the project exists and get its ID
		projectID := repoConfig.ProjectID
		if projectID == "" && repoConfig.ProjectName != "" {
			// Try to find project ID from name
			projects, err := client.ListProjects()
			if err != nil {
				fmt.Println("Error listing projects:", err)
				return nil
			}

			for _, p := range projects {
				if p.Name == repoConfig.ProjectName {
					projectID = p.ID

					// Save the project ID for future use
					repoConfig.ProjectID = projectID
					if err := config.SaveRepoConfig(repoConfig); err != nil {
						fmt.Println("Warning: Failed to save project ID to config:", err)
						// Continue anyway
					} else {
						fmt.Printf("Cached project ID: %s\n", projectID)
					}
					break
				}
			}

			if projectID == "" {
				fmt.Printf("Error: Could not find project with name: %s\n", repoConfig.ProjectName)
				return nil
			}
		}

		if projectID == "" {
			fmt.Println("Error: No project linked. Link a project first with 'hhx project link'")
			return nil
		}

		// Check if the remote collection exists
		fmt.Printf("Checking if collection '%s' exists on the remote server...\n", remoteBucketName)

		// Call the API to check collections
		collectionsURL := fmt.Sprintf("%s/%s/projects/%s/collections", remoteURL, api.API_VERSION, projectID)

		token, err := tokenStore.GetToken()
		if err != nil {
			fmt.Println("Error getting token:", err)
			return nil
		}

		req, err := http.NewRequest("GET", collectionsURL, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return nil
		}

		req.Header.Set("Authorization", "Bearer "+token)

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error checking remote collections:", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Failed to check collections with status %d: %s\n", resp.StatusCode, string(bodyBytes))
			return nil
		}

		var collectionsResponse struct {
			Collections []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"collections"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&collectionsResponse); err != nil {
			fmt.Println("Error parsing collections response:", err)
			return nil
		}

		// Check if our collection exists in the remote collections
		collectionExists := false
		for _, c := range collectionsResponse.Collections {
			if c.Name == remoteBucketName {
				collectionExists = true
				break
			}
		}

		if !collectionExists {
			if createIfMissing {
				fmt.Printf("Collection '%s' doesn't exist on the server. Creating it...\n", remoteBucketName)

				// Create the collection on the server
				createURL := fmt.Sprintf("%s/%s/projects/%s/collections", remoteURL, api.API_VERSION, projectID)

				createRequest := struct {
					Name string `json:"name"`
					Type string `json:"type"`
					Path string `json:"path"`
				}{
					Name: remoteBucketName,
					Type: string(collection.Type),
					Path: collection.Path,
				}

				jsonData, err := json.Marshal(createRequest)
				if err != nil {
					fmt.Println("Error preparing collection creation request:", err)
					return nil
				}

				req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Println("Error creating request:", err)
					return nil
				}

				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")

				resp, err := httpClient.Do(req)
				if err != nil {
					fmt.Println("Error creating remote collection:", err)
					return nil
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Printf("Failed to create collection with status %d: %s\n", resp.StatusCode, string(bodyBytes))
					return nil
				}

				fmt.Printf("Remote collection '%s' created successfully\n", remoteBucketName)
				collectionExists = true
			} else {
				fmt.Printf("Error: Collection '%s' does not exist on the remote server.\n", remoteBucketName)
				fmt.Println("Use --create flag to create it automatically, or create it manually on the server first.")
				return nil
			}
		}

		// Update the local collection with remote connection info
		if collection.Metadata == nil {
			collection.Metadata = make(map[string]interface{})
		}
		collection.Metadata["remoteName"] = remoteBucketName
		collection.Metadata["remoteLinkTime"] = time.Now().Format(time.RFC3339)

		// Save the updated collection
		if err := index.UpdateCollection(collection); err != nil {
			fmt.Println("Error updating collection metadata:", err)
			return nil
		}

		// Save index
		if err := index.Save(repoConfig.IndexPath); err != nil {
			fmt.Println("Error saving index:", err)
			return nil
		}

		fmt.Printf("Collection '%s' successfully linked to remote collection '%s'.\n",
			collectionName, remoteBucketName)
		fmt.Printf("You can now push files with: hhx push --collection=%s\n", collectionName)

		return nil
	},
}

func init() {
	collectionCmd.AddCommand(collectionLinkCmd)
	collectionLinkCmd.Flags().String("bucket", "", "Remote bucket name (defaults to local collection name)")
	collectionLinkCmd.Flags().Bool("create", false, "Create the remote collection if it doesn't exist")
}
