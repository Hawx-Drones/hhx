package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Index represents the staging area and metadata store
type Index struct {
	// Files in the staging area, keyed by their path
	Files map[string]*File `json:"files"`

	// Deleted files, keyed by their path
	Deleted map[string]*File `json:"deleted,omitempty"`

	// Previously synced files, keyed by their path
	Synced map[string]*File `json:"synced,omitempty"`

	// Collections available in this repository
	Collections map[string]*Collection `json:"collections,omitempty"`

	// Default collection to use when none is specified
	DefaultCollection string `json:"default_collection,omitempty"`

	// Repository root directory
	RepoRoot string `json:"repo_root"`

	// Mutex for concurrent access
	mu sync.RWMutex `json:"-"`
}

// NewIndex creates a new index
func NewIndex(repoRoot string) *Index {
	return &Index{
		Files:       make(map[string]*File),
		Deleted:     make(map[string]*File),
		Synced:      make(map[string]*File),
		Collections: make(map[string]*Collection),
		RepoRoot:    repoRoot,
	}
}

// Load loads the index from the given file
func LoadIndex(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Determine repository root from index path
			repoRoot := filepath.Dir(filepath.Dir(path))
			return NewIndex(repoRoot), nil
		}
		return nil, err
	}

	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	// Initialize maps if they're nil
	if index.Files == nil {
		index.Files = make(map[string]*File)
	}
	if index.Deleted == nil {
		index.Deleted = make(map[string]*File)
	}
	if index.Synced == nil {
		index.Synced = make(map[string]*File)
	}
	if index.Collections == nil {
		index.Collections = make(map[string]*Collection)
	}

	return &index, nil
}

// Save saves the index to the given file
func (idx *Index) Save(path string) error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// AddCollection adds a new collection to the index
func (idx *Index) AddCollection(collection *Collection) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Validate the collection
	if err := collection.Validate(); err != nil {
		return err
	}

	// Check if a collection with the same name already exists
	if _, exists := idx.Collections[collection.Name]; exists {
		return ErrCollectionExists
	}

	idx.Collections[collection.Name] = collection

	// If this is the first collection, set it as the default
	if len(idx.Collections) == 1 {
		idx.DefaultCollection = collection.Name
	}

	return nil
}

// UpdateCollection updates an existing collection
func (idx *Index) UpdateCollection(collection *Collection) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Validate the collection
	if err := collection.Validate(); err != nil {
		return err
	}

	// Check if the collection exists
	if _, exists := idx.Collections[collection.Name]; !exists {
		return ErrCollectionNotFound
	}

	idx.Collections[collection.Name] = collection
	return nil
}

// RemoveCollection removes a collection from the index
func (idx *Index) RemoveCollection(name string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Check if the collection exists
	if _, exists := idx.Collections[name]; !exists {
		return ErrCollectionNotFound
	}

	delete(idx.Collections, name)

	// If the default collection was removed, update the default
	if idx.DefaultCollection == name {
		if len(idx.Collections) > 0 {
			// Choose the first collection as the default
			for collName := range idx.Collections {
				idx.DefaultCollection = collName
				break
			}
		} else {
			idx.DefaultCollection = ""
		}
	}

	return nil
}

// GetCollection gets a collection by name
func (idx *Index) GetCollection(name string) (*Collection, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	collection, exists := idx.Collections[name]
	if !exists {
		return nil, ErrCollectionNotFound
	}

	return collection, nil
}

// GetCollections returns all collections
func (idx *Index) GetCollections() []*Collection {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	collections := make([]*Collection, 0, len(idx.Collections))
	for _, collection := range idx.Collections {
		collections = append(collections, collection)
	}

	return collections
}

// GetDefaultCollection gets the default collection
func (idx *Index) GetDefaultCollection() (*Collection, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if idx.DefaultCollection == "" {
		return nil, ErrNoDefaultCollection
	}

	collection, exists := idx.Collections[idx.DefaultCollection]
	if !exists {
		return nil, ErrCollectionNotFound
	}

	return collection, nil
}

// SetDefaultCollection sets the default collection
func (idx *Index) SetDefaultCollection(name string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.Collections[name]; !exists {
		return ErrCollectionNotFound
	}

	idx.DefaultCollection = name
	return nil
}

// StageFile stages a file
func (idx *Index) StageFile(path string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	file, err := NewFileFromPath(idx.RepoRoot, path)
	if err != nil {
		return err
	}

	if file == nil {
		// Skip directories
		return nil
	}

	// Check if this file was previously synced
	if synced, ok := idx.Synced[file.Path]; ok {
		if synced.Hash == file.Hash {
			file.Status = StatusSynced
			file.RemoteURL = synced.RemoteURL
		} else {
			file.Status = StatusModified
			file.RemoteURL = synced.RemoteURL
		}
	} else {
		file.Status = StatusUntracked
	}

	// Now that we've updated the status, mark it as staged
	file.Status = StatusStaged
	idx.Files[file.Path] = file

	// Remove from deleted if it was there
	delete(idx.Deleted, file.Path)

	return nil
}

// UnstageFile unstages a file
func (idx *Index) UnstageFile(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	relPath, err := filepath.Rel(idx.RepoRoot, path)
	if err != nil {
		return
	}
	relPath = filepath.ToSlash(relPath)

	delete(idx.Files, relPath)
}

// StageDirectory stages all files in a directory recursively
func (idx *Index) StageDirectory(dirPath string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip .git, .hhx directories and other hidden directories
			if info.Name() == ".git" || info.Name() == ".hhx" || info.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		file, err := NewFileFromPath(idx.RepoRoot, path)
		if err != nil {
			return err
		}

		if file == nil {
			return nil
		}

		// Check if this file was previously synced
		if synced, ok := idx.Synced[file.Path]; ok {
			if synced.Hash == file.Hash {
				file.Status = StatusSynced
				file.RemoteURL = synced.RemoteURL
			} else {
				file.Status = StatusModified
				file.RemoteURL = synced.RemoteURL
			}
		} else {
			file.Status = StatusUntracked
		}

		// Now that we've updated the status, mark it as staged
		file.Status = StatusStaged
		idx.Files[file.Path] = file

		return nil
	})
}

// MarkSynced marks a file as synced
func (idx *Index) MarkSynced(path string, remoteURL string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if file, ok := idx.Files[path]; ok {
		file.Status = StatusSynced
		file.RemoteURL = remoteURL
		idx.Synced[path] = file
		delete(idx.Files, path)
	}
}

// GetStagedFiles returns all staged files
func (idx *Index) GetStagedFiles() []*File {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	files := make([]*File, 0, len(idx.Files))
	for _, file := range idx.Files {
		files = append(files, file)
	}
	return files
}

// GetAllFiles returns all tracked files (staged, synced, and deleted)
func (idx *Index) GetAllFiles() []*File {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	files := make([]*File, 0, len(idx.Files)+len(idx.Synced)+len(idx.Deleted))

	for _, file := range idx.Files {
		files = append(files, file)
	}

	for _, file := range idx.Synced {
		files = append(files, file)
	}

	for _, file := range idx.Deleted {
		files = append(files, file)
	}

	return files
}

// ScanWorkingDirectory scans the working directory for changes
// ScanWorkingDirectory scans the working directory for changes
func (idx *Index) ScanWorkingDirectory() ([]*File, []*File, []*File, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Copy synced files to a new map to track which ones we've seen
	seen := make(map[string]bool)
	for path := range idx.Synced {
		seen[path] = false
	}

	// Track new, modified, and unchanged files
	var newFiles, modifiedFiles, unchangedFiles []*File

	err := filepath.Walk(idx.RepoRoot, func(path string, info os.FileInfo, err error) error {
		// Handle errors from filepath.Walk
		if err != nil {
			// Skip files that can't be accessed instead of stopping the entire walk
			return nil
		}

		// Skip if info is nil
		if info == nil {
			return nil
		}

		// Skip .git, .hhx directories and other hidden directories
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == ".hhx" || info.Name()[0] == '.' {
				return filepath.SkipDir
			}

			// Skip build and compilation directories
			if info.Name() == "cmake-build-debug" || info.Name() == "build" {
				return filepath.SkipDir
			}

			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(idx.RepoRoot, path)
		if err != nil {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		// Skip any file in the .hhx directory
		if strings.HasPrefix(relPath, ".hhx/") {
			return nil
		}

		// Calculate hash
		hash, err := hashFile(path)
		if err != nil {
			// Skip files that can't be hashed
			return nil
		}

		// Check if this file was previously synced
		if synced, ok := idx.Synced[relPath]; ok {
			seen[relPath] = true

			if synced.Hash != hash {
				// File was modified
				file := &File{
					Path:         relPath,
					Size:         info.Size(),
					Hash:         hash,
					LastModified: info.ModTime(),
					Status:       StatusModified,
					RemoteURL:    synced.RemoteURL,
				}
				modifiedFiles = append(modifiedFiles, file)
			} else {
				// File is unchanged
				unchangedFiles = append(unchangedFiles, synced)
			}
		} else {
			// New file
			file := &File{
				Path:         relPath,
				Size:         info.Size(),
				Hash:         hash,
				LastModified: info.ModTime(),
				Status:       StatusUntracked,
			}
			newFiles = append(newFiles, file)
		}

		return nil
	})

	fmt.Printf("err: %v\n", err)

	// Find deleted files
	var deletedFiles []*File
	for path, wasSeen := range seen {
		if !wasSeen {
			// File was deleted
			file := idx.Synced[path]
			file.Status = StatusUntracked
			deletedFiles = append(deletedFiles, file)

			// Add to deleted files list
			idx.Deleted[path] = file

			// Remove from synced files
			delete(idx.Synced, path)
		}
	}

	return newFiles, modifiedFiles, deletedFiles, nil
}
