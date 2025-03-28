package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client handles communication with the API server
type Client struct {
	// Base URL of the API server
	BaseURL string

	// Authentication token
	AuthToken string

	// HTTP client with a timeout
	client *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL, authToken string) *Client {
	return &Client{
		BaseURL:   baseURL,
		AuthToken: authToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Auth contains authentication response
type Auth struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// PushResponse contains the response from the server after pushing files
type PushResponse struct {
	UploadedFiles []UploadedFile `json:"uploaded_files"`
	Errors        []UploadError  `json:"errors,omitempty"`
}

// UploadedFile contains information about an uploaded file
type UploadedFile struct {
	Path      string `json:"path"`
	RemoteURL string `json:"remote_url"`
	Size      int64  `json:"size"`
	Hash      string `json:"hash"`
}

// UploadError contains information about a failed upload
type UploadError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

// CollectionInfo contains information about a collection for API requests
type CollectionInfo struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Path     string                 `json:"path"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TODO: This file is generic for now and will be updated in future PRs
// Login authenticates the user with the server
func (c *Client) Login(email, password string) (*Auth, error) {
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/login", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s (%d)", string(body), resp.StatusCode)
	}

	var auth Auth
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return nil, err
	}

	// Update the client's auth token
	c.AuthToken = auth.Token

	return &auth, nil
}

// PushFilesToCollection uploads files to a specific collection
func (c *Client) PushFilesToCollection(repoRoot string, files []*models.File, collection *models.Collection) (*PushResponse, error) {
	var response PushResponse

	// Use a multipart form to upload multiple files
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add metadata about the files
	filesMeta, err := json.Marshal(files)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteField("files_meta", string(filesMeta)); err != nil {
		return nil, err
	}

	// Add collection information
	collectionInfo := CollectionInfo{
		Name: collection.Name,
		Type: string(collection.Type),
		Path: collection.Path,
	}

	// Convert schema to a generic map if it exists
	if collection.Schema != nil {
		schemaBytes, err := json.Marshal(collection.Schema)
		if err != nil {
			return nil, err
		}

		var schemaMap map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return nil, err
		}

		collectionInfo.Schema = schemaMap
	}

	// Add collection metadata if it exists
	if collection.Metadata != nil {
		collectionInfo.Metadata = collection.Metadata
	}

	collectionJSON, err := json.Marshal(collectionInfo)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteField("collection", string(collectionJSON)); err != nil {
		return nil, err
	}

	// Add each file
	for _, file := range files {
		filePath := filepath.Join(repoRoot, filepath.FromSlash(file.Path))
		f, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		part, err := writer.CreateFormFile("file", file.Path)
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(part, f); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/files/upload", c.BaseURL), body)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed: %s (%d)", string(responseBody), resp.StatusCode)
	}

	// Parse response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(collection *models.Collection) error {
	// Convert collection to CollectionInfo
	collectionInfo := CollectionInfo{
		Name: collection.Name,
		Type: string(collection.Type),
		Path: collection.Path,
	}

	// Convert schema to a generic map if it exists
	if collection.Schema != nil {
		schemaBytes, err := json.Marshal(collection.Schema)
		if err != nil {
			return err
		}

		var schemaMap map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &schemaMap); err != nil {
			return err
		}

		collectionInfo.Schema = schemaMap
	}

	// Add collection metadata if it exists
	if collection.Metadata != nil {
		collectionInfo.Metadata = collection.Metadata
	}

	reqBody, err := json.Marshal(collectionInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/collections", c.BaseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create collection failed: %s (%d)", string(responseBody), resp.StatusCode)
	}

	return nil
}

// GetStatus gets the status of files from the server
func (c *Client) GetStatus() ([]*models.File, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/files/status", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get status: %s (%d)", string(body), resp.StatusCode)
	}

	var files []*models.File
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

// ListCollections gets the list of available collections from the server
func (c *Client) ListCollections() ([]*models.Collection, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/collections", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.AuthToken))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list collections: %s (%d)", string(body), resp.StatusCode)
	}

	var collectionsInfo []CollectionInfo
	if err := json.NewDecoder(resp.Body).Decode(&collectionsInfo); err != nil {
		return nil, err
	}

	// Convert CollectionInfo to Collection
	collections := make([]*models.Collection, 0, len(collectionsInfo))
	for _, info := range collectionsInfo {
		collection := &models.Collection{
			Name: info.Name,
			Type: models.CollectionType(info.Type),
			Path: info.Path,
		}

		// Convert schema if it exists
		if info.Schema != nil {
			schemaBytes, err := json.Marshal(info.Schema)
			if err != nil {
				return nil, err
			}

			var schema models.Schema
			if err := json.Unmarshal(schemaBytes, &schema); err != nil {
				return nil, err
			}

			collection.Schema = &schema
		}

		// Add metadata if it exists
		if info.Metadata != nil {
			collection.Metadata = info.Metadata
		}

		collections = append(collections, collection)
	}

	return collections, nil
}
