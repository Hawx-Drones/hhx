package models

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileStatus represents the status of a file
type FileStatus string

const (
	StatusUntracked FileStatus = "untracked" // File is not tracked
	StatusModified  FileStatus = "modified"  // File is tracked but modified
	StatusStaged    FileStatus = "staged"    // File is staged for commit
	StatusSynced    FileStatus = "synced"    // File is synced with the server
)

// File represents a file in the repository
type File struct {
	Path         string     `json:"path"`                 // Relative path from repository root
	Size         int64      `json:"size"`                 // File size in bytes
	Hash         string     `json:"hash"`                 // SHA-256 hash of file content
	LastModified time.Time  `json:"last_modified"`        // Last modification time
	Status       FileStatus `json:"status"`               // File status
	RemoteURL    string     `json:"remote_url,omitempty"` // URL of the file on the server
	Collection   string     `json:"collection,omitempty"` // Collection name
}

// NewFileFromPath creates a new File instance from the given path
func NewFileFromPath(repoRoot, path string) (*File, error) {
	relativePath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return nil, err
	}

	// Use forward slashes for consistency across platforms
	relativePath = filepath.ToSlash(relativePath)

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Skip directories
	if info.IsDir() {
		return nil, nil
	}

	// Calculate hash
	hash, err := hashFile(path)
	if err != nil {
		return nil, err
	}

	file := &File{
		Path:         relativePath,
		Size:         info.Size(),
		Hash:         hash,
		LastModified: info.ModTime(),
		Status:       StatusUntracked,
	}

	return file, nil
}

// hashFile calculates the SHA-256 hash of a file
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// FullPath returns the full path of the file
func (f *File) FullPath(repoRoot string) string {
	return filepath.Join(repoRoot, filepath.FromSlash(f.Path))
}
