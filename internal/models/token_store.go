package models

import (
	"os"
	"path/filepath"
)

type TokenStore struct {
	TokenFile string
}

func NewTokenStore(configDir string) *TokenStore {
	return &TokenStore{
		TokenFile: filepath.Join(configDir, ".auth_token"),
	}
}

func (ts *TokenStore) SaveToken(token string) error {
	return os.WriteFile(ts.TokenFile, []byte(token), 0600) // Restricted permissions
}

func (ts *TokenStore) GetToken() (string, error) {
	data, err := os.ReadFile(ts.TokenFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (ts *TokenStore) ClearToken() error {
	if _, err := os.Stat(ts.TokenFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clear
	}
	return os.Remove(ts.TokenFile)
}
