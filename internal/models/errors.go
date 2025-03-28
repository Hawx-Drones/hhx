package models

import (
	"errors"
)

// Collection-related errors
var (
	// ErrCollectionExists is returned when trying to create a collection that already exists
	ErrCollectionExists = errors.New("collection already exists")

	// ErrCollectionNotFound is returned when a collection is not found
	ErrCollectionNotFound = errors.New("collection not found")

	// ErrNoDefaultCollection is returned when no default collection is set
	ErrNoDefaultCollection = errors.New("no default collection set")

	// ErrInvalidCollectionType is returned when an invalid collection type is specified
	ErrInvalidCollectionType = errors.New("invalid collection type")

	// ErrSchemaRequired is returned when a schema is required but not provided
	ErrSchemaRequired = errors.New("schema is required for this collection type")

	// ErrInvalidSchema is returned when the schema is invalid
	ErrInvalidSchema = errors.New("invalid schema")
)

// File-related errors
var (
	// ErrFileNotFound is returned when a file is not found
	ErrFileNotFound = errors.New("file not found")

	// ErrFileAlreadyExists is returned when a file already exists
	ErrFileAlreadyExists = errors.New("file already exists")
)
