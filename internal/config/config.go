package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// API server URL
	ServerURL string `json:"server_url"`

	// Authentication token (will be populated after login)
	AuthToken string `json:"auth_token,omitempty"`

	// User information
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`

	// Default local repository path
	DefaultRepoPath string `json:"default_repo_path,omitempty"`
}

// Load loads the configuration from the given file path
func Load(path string) (*Config, error) {
	// If config file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{
			ServerURL: "http://localhost:8080", // Default local server URL
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the given file path
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetRepoConfigPath returns the path to the repository configuration file
func GetRepoConfigPath() (string, error) {
	// Start from current directory and traverse up until we find .hhx directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		hhxDir := filepath.Join(dir, ".hhx")
		if _, err := os.Stat(hhxDir); err == nil {
			return filepath.Join(hhxDir, "config.json"), nil
		}

		// Stop if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// RepoConfig represents repository-specific configuration
type RepoConfig struct {
	// Remote name and URL mapping
	Remotes map[string]string `json:"remotes"`

	// Current remote name
	CurrentRemote string `json:"current_remote"`

	// Index file path
	IndexPath string `json:"index_path"`
}

// LoadRepoConfig loads the repository configuration
func LoadRepoConfig() (*RepoConfig, error) {
	path, err := GetRepoConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg RepoConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveRepoConfig saves the repository configuration
func SaveRepoConfig(cfg *RepoConfig) error {
	path, err := GetRepoConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
