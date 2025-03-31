package models

// Bucket represents a storage bucket
type Bucket struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	ProjectID        string   `json:"project_id"`
	CreatorID        string   `json:"creator_id"`
	Public           bool     `json:"public"`
	AllowedFileTypes []string `json:"allowed_file_types,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}
