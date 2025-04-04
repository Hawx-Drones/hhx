package models

import (
	"encoding/json"
	"fmt"
)

// CollectionType defines the type of collection
type CollectionType string

const (
	// CollectionTypeBucket represents an object storage bucket
	CollectionTypeBucket CollectionType = "bucket"

	// CollectionTypeTable represents a database table
	CollectionTypeTable CollectionType = "table"
)

// Collection represents a destination for data (bucket or table)
type Collection struct {
	// Name of the collection
	Name string `json:"name"`

	// Type of collection (bucket or table)
	Type CollectionType `json:"type"`

	// Path within the remote (e.g., "models/" for a bucket or a table name)
	Path string `json:"path"`

	// Schema definition for tables (nil for buckets)
	Schema *Schema `json:"schema,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Schema represents the structure of a table
type Schema struct {
	// Columns in the table
	Columns []*Column `json:"columns"`
}

// Column represents a column in a table schema
type Column struct {
	// Name of the column
	Name string `json:"name"`

	// Data type of the column
	Type string `json:"type"`

	// Whether this column is a primary key
	PrimaryKey bool `json:"primary_key,omitempty"`

	// Whether this column can be null
	Nullable bool `json:"nullable,omitempty"`

	// Default value for the column
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// CollectionsResponse represents the response when listing collections
type CollectionsResponse struct {
	Collections []struct {
		Name string `json:"name"`
	} `json:"collections"`
}

// Validate validates the collection
func (c *Collection) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	if c.Type != CollectionTypeBucket && c.Type != CollectionTypeTable {
		return fmt.Errorf("invalid collection type: %s", c.Type)
	}

	if c.Path == "" {
		return fmt.Errorf("collection path cannot be empty")
	}

	// For tables, schema is required
	if c.Type == CollectionTypeTable && c.Schema == nil {
		return fmt.Errorf("schema is required for table collections")
	}

	// For buckets, schema should be nil
	if c.Type == CollectionTypeBucket && c.Schema != nil {
		return fmt.Errorf("schema is not applicable for bucket collections")
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (c *Collection) MarshalJSON() ([]byte, error) {
	type Alias Collection
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	})
}
