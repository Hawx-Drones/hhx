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
	if err := validatePushInputs(projectNameOrID, collection); err != nil {
		return nil, err
	}

	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	projectID, err := c.resolveProjectID(projectNameOrID, token)
	if err != nil {
		return nil, err
	}

	if err := c.verifyCollectionExists(projectID, collection.Name, token); err != nil {
		return nil, err
	}

	requestBody, contentType, err := c.createMultipartRequest(repoRoot, files, collection)
	if err != nil {
		return nil, err
	}

	return c.sendPushRequest(projectID, collection.Name, requestBody, contentType, token)
}

// validatePushInputs validates the inputs for the push operation
func validatePushInputs(projectNameOrID string, collection *models.Collection) error {
	if projectNameOrID == "" {
		return fmt.Errorf("project name or ID is required")
	}

	if collection == nil {
		return fmt.Errorf("collection cannot be nil")
	}

	if collection.Name == "" {
		return fmt.Errorf("collection name is required")
	}

	return nil
}

// resolveProjectID resolves a project name to its ID
func (c *Client) resolveProjectID(projectNameOrID string, token string) (string, error) {
	if util.IsUUID(projectNameOrID) {
		return projectNameOrID, nil
	}

	// Not a UUID, try to get the project by name
	fmt.Printf("Project '%s' doesn't look like a UUID, looking up project ID...\n", projectNameOrID)
	projects, err := c.ListProjects()
	if err != nil {
		return "", fmt.Errorf("error listing projects: %w", err)
	}

	for _, p := range projects {
		if p.Name == projectNameOrID {
			fmt.Printf("Found project ID: %s\n", p.ID)
			return p.ID, nil
		}
	}

	return "", fmt.Errorf("project not found with name: %s", projectNameOrID)
}

// verifyCollectionExists checks if a collection exists on the server
func (c *Client) verifyCollectionExists(projectID, collectionName, token string) error {
	collectionsURL := fmt.Sprintf("%s/%s/projects/%s/collections", c.BaseURL, API_VERSION, projectID)

	req, err := http.NewRequest("GET", collectionsURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request to check collections: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error checking collections: %w", err)
	}
	defer safelyCloseResponseBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to check collections with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var collectionsResponse models.CollectionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&collectionsResponse); err != nil {
		return fmt.Errorf("error decoding collections response: %w", err)
	}

	// Check if the collection exists in the remote collections
	for _, c := range collectionsResponse.Collections {
		if c.Name == collectionName {
			return nil // Collection exists
		}
	}

	return fmt.Errorf("collection '%s' does not exist on the remote server. Please create it first using API calls", collectionName)
}

// createMultipartRequest creates a multipart request with files and metadata
func (c *Client) createMultipartRequest(repoRoot string, files []*models.File, collection *models.Collection) (*bytes.Buffer, string, error) {
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if err := addMetadataToRequest(writer, collection); err != nil {
		return nil, "", err
	}

	// Add each file to the form
	for _, file := range files {
		if err := addFileToRequest(writer, repoRoot, file); err != nil {
			return nil, "", err
		}
	}

	contentType := writer.FormDataContentType()
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	return &requestBody, contentType, nil
}

// addMetadataToRequest adds collection metadata to the multipart request
func addMetadataToRequest(writer *multipart.Writer, collection *models.Collection) error {
	metadataField, err := writer.CreateFormField("metadata")
	if err != nil {
		return fmt.Errorf("error creating metadata field: %w", err)
	}

	metadata := map[string]interface{}{
		"collection_type": collection.Type,
		"collection_path": collection.Path,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("error marshalling metadata: %w", err)
	}

	if _, err := metadataField.Write(metadataBytes); err != nil {
		return fmt.Errorf("error writing metadata: %w", err)
	}

	return nil
}

// addFileToRequest adds a file and its metadata to the multipart request
func addFileToRequest(writer *multipart.Writer, repoRoot string, file *models.File) error {
	fullPath := filepath.Join(repoRoot, file.Path)
	f, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", file.Path, err)
	}
	defer safelyCloseFile(f)

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("error getting file stats for %s: %w", file.Path, err)
	}

	contentType := getContentTypeFromFilename(file.Path)
	fileField, err := writer.CreateFormFile("files", file.Path)
	if err != nil {
		return fmt.Errorf("error creating form file: %w", err)
	}

	if _, err := io.Copy(fileField, f); err != nil {
		return fmt.Errorf("error copying file data: %w", err)
	}

	return addFileMetadataToRequest(writer, file, stat, contentType)
}

// addFileMetadataToRequest adds file metadata to the multipart request
func addFileMetadataToRequest(writer *multipart.Writer, file *models.File, stat os.FileInfo, contentType string) error {
	fileMetaField, err := writer.CreateFormField(fmt.Sprintf("file_meta_%s", file.Path))
	if err != nil {
		return fmt.Errorf("error creating file metadata field: %w", err)
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
		return fmt.Errorf("error marshalling file metadata: %w", err)
	}

	if _, err := fileMetaField.Write(fileMetaBytes); err != nil {
		return fmt.Errorf("error writing file metadata: %w", err)
	}

	return nil
}

// sendPushRequest sends the push request to the server
func (c *Client) sendPushRequest(projectID, collectionName string, requestBody *bytes.Buffer, contentType, token string) (*PushResponse, error) {
	url := fmt.Sprintf("%s/%s/projects/%s/collections/%s/files", c.BaseURL, API_VERSION, projectID, collectionName)

	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer safelyCloseResponseBody(resp.Body)

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

// safelyCloseResponseBody safely closes a response body
func safelyCloseResponseBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		fmt.Printf("Warning: Failed to close response body: %v\n", err)
	}
}

// safelyCloseFile safely closes a file
func safelyCloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Printf("Warning: Failed to close file handle: %v\n", err)
	}
}

// getContentTypeFromFilename returns the content type based on the file extension
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
