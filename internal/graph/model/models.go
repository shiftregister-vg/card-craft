package model

import "github.com/shiftregister-vg/card-craft/internal/models"

// AuthPayload represents the response from authentication operations
type AuthPayload struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// CardSearchResult represents the paginated result of a card search
type CardSearchResult struct {
	Cards      []*models.Card `json:"cards"`
	TotalCount int            `json:"totalCount"`
	Page       int            `json:"page"`
	PageSize   int            `json:"pageSize"`
}
