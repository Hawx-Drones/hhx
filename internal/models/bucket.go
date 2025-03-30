package models

// Bucket represents a storage bucket
type Bucket struct {
	Name      string `json:"name"`
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Public    bool   `json:"public"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
