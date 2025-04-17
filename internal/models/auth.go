package models

// AuthPayload represents the response for authentication operations
type AuthPayload struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
