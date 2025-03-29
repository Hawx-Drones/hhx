package models

// UserDetails represents basic information about a user
type UserDetails struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"password,omitempty"`
}

// UserDetailsWithSubscription represents detailed information about a user, including subscription details
type UserDetailsWithSubscription struct {
	UserID       string       `json:"id"`
	Email        string       `json:"email"`
	Name         string       `json:"name"`
	Phone        string       `json:"phone"`
	Subscription Subscription `json:"subscription,omitempty"`
}
