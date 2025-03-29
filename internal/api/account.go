package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"io"
	"net/http"
)

// GetUserDetails fetches detailed information about the current user
func (c *Client) GetUserDetails() (*models.UserDetailsWithSubscription, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	url := fmt.Sprintf("%s/%s/account/me", c.BaseURL, API_VERSION)

	req, err := http.NewRequest("GET", url, nil)
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
		responseBody := string(bodyBytes)
		return nil, fmt.Errorf("failed to fetch user details with status %d: %s", resp.StatusCode, responseBody)
	}

	var response struct {
		User models.UserDetailsWithSubscription `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response.User, nil
}

// UpdateAccount updates the user's account information
func (c *Client) UpdateAccount(req *models.UserDetails) (*models.UserDetails, error) {
	token, err := c.tokenStore.GetToken()
	if err != nil {
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	url := fmt.Sprintf("%s/%s/account/update", c.BaseURL, API_VERSION)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
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
		return nil, fmt.Errorf("account update failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Message string             `json:"message"`
		User    models.UserDetails `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &response.User, nil
}
