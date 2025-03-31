package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"io"
	"net/http"
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

	// Token store for managing authentication tokens
	tokenStore *models.TokenStore
}

// NewClient creates a new API client
func NewClient(baseURL string, tokenStore *models.TokenStore) *Client {
	token := ""
	if tokenStore != nil {
		storedToken, err := tokenStore.GetToken()
		if err == nil && storedToken != "" {
			token = storedToken
		}
	}

	return &Client{
		BaseURL:    baseURL,
		AuthToken:  token,
		tokenStore: tokenStore,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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

// CreateCollection creates a new collection
func (c *Client) CreateCollection(collection *models.Collection) error {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list collections: %s (%d)", string(body), resp.StatusCode)
	}

	var collectionsInfo []CollectionInfo
	if err := json.NewDecoder(resp.Body).Decode(&collectionsInfo); err != nil {
		return nil, err
	}

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

		if info.Metadata != nil {
			collection.Metadata = info.Metadata
		}

		collections = append(collections, collection)
	}

	return collections, nil
}
