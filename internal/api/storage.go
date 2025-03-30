package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"io"
	"net/http"
	"net/url"
)

// ListBuckets retrieves all storage buckets
func (c *Client) ListBuckets() ([]models.Bucket, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	bucketsUrl := fmt.Sprintf("%s/%s/storage/buckets", c.BaseURL, API_VERSION)

	req, err := http.NewRequest("GET", bucketsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
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
		return nil, fmt.Errorf("failed to list buckets with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Buckets []models.Bucket `json:"buckets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Buckets, nil
}

// CreateBucket creates a new storage bucket
func (c *Client) CreateBucket(name string, public bool, fileSizeLimit int64) (*models.Bucket, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	bucketsUrl := fmt.Sprintf("%s/%s/storage/buckets", c.BaseURL, API_VERSION)

	request := struct {
		Name          string `json:"name"`
		Public        bool   `json:"public"`
		FileSizeLimit int64  `json:"fileSizeLimit,omitempty"`
	}{
		Name:          name,
		Public:        public,
		FileSizeLimit: fileSizeLimit,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", bucketsUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bucket creation failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Bucket models.Bucket `json:"bucket"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response.Bucket, nil
}

// GetBucket retrieves details about a specific bucket
func (c *Client) GetBucket(name string) (*models.Bucket, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	encodedName := url.PathEscape(name)
	bucketsUrl := fmt.Sprintf("%s/%s/storage/buckets/%s", c.BaseURL, API_VERSION, encodedName)

	req, err := http.NewRequest("GET", bucketsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
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
		return nil, fmt.Errorf("failed to get bucket with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Bucket models.Bucket `json:"bucket"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response.Bucket, nil
}

// UpdateBucket updates a bucket's settings
func (c *Client) UpdateBucket(name string, updates interface{}) error {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	encodedName := url.PathEscape(name)
	url := fmt.Sprintf("%s/%s/storage/buckets/%s", c.BaseURL, API_VERSION, encodedName)

	jsonData, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bucket update failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DeleteBucket deletes a storage bucket
func (c *Client) DeleteBucket(name string) error {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	encodedName := url.PathEscape(name)
	url := fmt.Sprintf("%s/%s/storage/buckets/%s", c.BaseURL, API_VERSION, encodedName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bucket deletion failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// EmptyBucket removes all files from a bucket
func (c *Client) EmptyBucket(name string) error {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return fmt.Errorf("error getting token: %w", err)
	}

	encodedName := url.PathEscape(name)
	url := fmt.Sprintf("%s/%s/storage/buckets/%s/empty", c.BaseURL, API_VERSION, encodedName)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("emptying bucket failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
