package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"hhx/internal/util"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PushResponse represents the response from a push operation
type PushResponse struct {
	UploadedFiles []struct {
		Path      string `json:"path"`
		RemoteURL string `json:"remote_url"`
		Size      int64  `json:"size"`
	} `json:"uploaded_files"`
	Errors []UploadError `json:"errors"`
}

// PushFilesToProjectCollection uploads files to a specific project and collection
func (c *Client) PushFilesToProjectCollection(repoRoot string, files []*models.File, projectNameOrID string, collection *models.Collection) (*PushResponse, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	// Validate inputs
	if projectNameOrID == "" {
		return nil, fmt.Errorf("project name or ID is required")
	}

	if collection == nil {
		return nil, fmt.Errorf("collection cannot be nil")
	}

	if collection.Name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	// Check if the input is already a UUID
	projectID := projectNameOrID
	if !util.IsUUID(projectNameOrID) {
		// Not a UUID, so try to get the project by name
		fmt.Printf("Project '%s' doesn't look like a UUID, looking up project ID...\n", projectNameOrID)
		projects, err := c.ListProjects()
		if err != nil {
			return nil, fmt.Errorf("error listing projects: %w", err)
		}

		found := false
		for _, p := range projects {
			if p.Name == projectNameOrID {
				projectID = p.ID
				fmt.Printf("Found project ID: %s\n", projectID)
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("project not found with name: %s", projectNameOrID)
		}
	}

	collectionsURL := fmt.Sprintf("%s/%s/projects/%s/collections", c.BaseURL, API_VERSION, projectID)

	req, err := http.NewRequest("GET", collectionsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to check collections: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error checking collections: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to check collections with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var collectionsResponse struct {
		Collections []struct {
			Name string `json:"name"`
		} `json:"collections"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&collectionsResponse); err != nil {
		return nil, fmt.Errorf("error decoding collections response: %w", err)
	}

	// Check if our collection exists in the remote collections
	collectionExists := false
	for _, c := range collectionsResponse.Collections {
		if c.Name == collection.Name {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		return nil, fmt.Errorf("collection '%s' does not exist on the remote server. Please create it first using API calls", collection.Name)
	}

	// Construct the URL with project ID and collection details
	url := fmt.Sprintf("%s/%s/projects/%s/collections/%s/files", c.BaseURL, API_VERSION, projectID, collection.Name)

	// Create a buffer to write our multipart form to
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add metadata about the push operation
	metadataField, err := writer.CreateFormField("metadata")
	if err != nil {
		return nil, fmt.Errorf("error creating metadata field: %w", err)
	}

	metadata := map[string]interface{}{
		"collection_type": collection.Type,
		"collection_path": collection.Path,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("error marshalling metadata: %w", err)
	}

	if _, err := metadataField.Write(metadataBytes); err != nil {
		return nil, fmt.Errorf("error writing metadata: %w", err)
	}

	// Add each file to the form
	for _, file := range files {
		fullPath := filepath.Join(repoRoot, file.Path)
		f, err := os.Open(fullPath)
		if err != nil {
			return nil, fmt.Errorf("error opening file %s: %w", file.Path, err)
		}

		stat, err := f.Stat()
		if err != nil {
			err := f.Close()
			if err != nil {
				return nil, fmt.Errorf("Warning: Failed to close file handle: %v\n", err)
			} // Close file handle on error
			return nil, fmt.Errorf("error getting file stats for %s: %w", file.Path, err)
		}

		contentType := getContentTypeFromFilename(file.Path)
		fileField, err := writer.CreateFormFile("files", file.Path)
		if err != nil {
			err := f.Close()
			if err != nil {
				return nil, fmt.Errorf("Warning: Failed to close file handle: %v\n", err)
			} // Close file handle on error
			return nil, fmt.Errorf("error creating form file: %w", err)
		}

		if _, err := io.Copy(fileField, f); err != nil {
			err := f.Close()
			if err != nil {
				return nil, fmt.Errorf("Warning: Failed to close file handle: %v\n", err)
			} // Close file handle on error
			return nil, fmt.Errorf("error copying file data: %w", err)
		}

		if err := f.Close(); err != nil {
			return nil, fmt.Errorf("error closing file: %w", err)
		}

		fileMetaField, err := writer.CreateFormField(fmt.Sprintf("file_meta_%s", file.Path))
		if err != nil {
			return nil, fmt.Errorf("error creating file metadata field: %w", err)
		}

		fileMeta := map[string]interface{}{
			"path":         file.Path,
			"size":         stat.Size(),
			"hash":         file.Hash,
			"remote_url":   file.RemoteURL,
			"content_type": contentType,
		}

		fileMetaBytes, err := json.Marshal(fileMeta)
		if err != nil {
			return nil, fmt.Errorf("error marshalling file metadata: %w", err)
		}

		if _, err := fileMetaField.Write(fileMetaBytes); err != nil {
			return nil, fmt.Errorf("error writing file metadata: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err = http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	client = &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var pushResponse PushResponse
	if err := json.NewDecoder(resp.Body).Decode(&pushResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &pushResponse, nil
}

func getContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	// add more content types as needed
	default:
		return "application/octet-stream"
	}
}
