package models

// Subscription represents a user's subscription information
type Subscription struct {
	Plan        string `json:"plan"`
	Status      string `json:"status"`
	RenewalDate string `json:"renewalDate"`
	MemberSince string `json:"memberSince"`
}
