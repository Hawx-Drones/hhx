package models

// Auth contains authentication response
type Auth struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}
