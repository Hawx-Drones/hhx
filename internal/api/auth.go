package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhx/internal/models"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

// Register creates a new user account and returns authentication information
func (c *Client) Register(email, password string) (*models.Auth, error) {
	url := fmt.Sprintf("%s/%s/auth/signup", c.BaseURL, API_VERSION)

	requestBody := map[string]string{
		"email":    email,
		"password": password,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Message string `json:"message"`
		User    struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if err := c.tokenStore.SaveToken(response.Token); err != nil {
		return nil, fmt.Errorf("error saving token: %w", err)
	}

	return &models.Auth{
		Token:  response.Token,
		UserID: response.User.ID,
		Email:  response.User.Email,
	}, nil
}

// Login authenticates the user with the server
func (c *Client) Login(email, password string) (*models.Auth, error) {
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/auth/signin", c.BaseURL, API_VERSION), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client, err := createHTTPClientWithCookieJar()
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
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
		return nil, fmt.Errorf("login failed: %s (%d)", string(body), resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseMap); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %w", err)
	}

	authResponse := &models.Auth{}
	authResponse.UserID, authResponse.Email = extractUserInfo(responseMap)
	authResponse.Token = findAuthToken(resp.Cookies(), responseMap)

	if authResponse.Token == "" {
		return nil, fmt.Errorf("no authentication token found in server response")
	}

	c.AuthToken = authResponse.Token
	if c.tokenStore != nil {
		if err := c.tokenStore.SaveToken(authResponse.Token); err != nil {
			return nil, fmt.Errorf("failed to save auth token: %w", err)
		}
	}

	return authResponse, nil
}

// Logout clears the authentication token and notifies the server
func (c *Client) Logout() error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/auth/signout", c.BaseURL, API_VERSION), nil)
	if err != nil {
		return fmt.Errorf("error creating logout request: %w", err)
	}

	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	client, err := createHTTPClientWithCookieJar()
	if err != nil {
		return fmt.Errorf("error creating HTTP client: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error calling signout endpoint: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Warning: Failed to close response body: %v\n", err)
		}
	}(resp.Body)

	// If the server returns a non-200 status code, log the error but continue with local logout
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Warning: Server signout returned status %d: %s\n", resp.StatusCode, string(body))
	}

	c.AuthToken = ""
	if c.tokenStore != nil {
		return c.tokenStore.ClearToken()
	}

	return nil
}

// createHTTPClientWithCookieJar creates an HTTP client with a cookie jar
func createHTTPClientWithCookieJar() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar: %w", err)
	}

	return &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}, nil
}

// extractUserInfo extracts user information from the response
func extractUserInfo(responseMap map[string]interface{}) (string, string) {
	userID, email := "", ""

	if userObj, ok := responseMap["user"].(map[string]interface{}); ok {
		// Extract user ID from various possible fields
		for _, field := range []string{"id", "_id", "uid"} {
			if id, ok := userObj[field].(string); ok {
				userID = id
				break
			}
		}

		if userEmail, ok := userObj["email"].(string); ok {
			email = userEmail
		}
	}

	return userID, email
}

// findAuthToken looks for an authentication token in cookies and response body
func findAuthToken(cookies []*http.Cookie, responseMap map[string]interface{}) string {
	for _, cookie := range cookies {
		if cookie.Value != "" {
			cookieName := strings.ToLower(cookie.Name)

			switch {
			case cookie.Name == "auth",
				cookie.Name == "token",
				cookie.Name == "jwt",
				cookie.Name == "session",
				strings.Contains(cookieName, "auth"),
				strings.Contains(cookieName, "token"),
				strings.Contains(cookieName, "session"),
				strings.Contains(cookieName, "jwt"):

				return cookie.Value
			}
		}
	}

	// If not found in cookies, check response body
	if token, ok := responseMap["token"].(string); ok {
		return token
	}

	return ""
}
