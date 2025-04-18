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

// CollectionInput represents the input for creating or updating a collection
type CollectionInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Game        string  `json:"game"`
}

// CollectionCardInput represents the input for adding or updating a card in a collection
type CollectionCardInput struct {
	CardID    string  `json:"cardId"`
	Quantity  int     `json:"quantity"`
	Condition *string `json:"condition,omitempty"`
	IsFoil    *bool   `json:"isFoil,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

// ImportSource represents the source and format for importing a collection
type ImportSource struct {
	Source string `json:"source"`
	Format string `json:"format"`
}

// ImportResult represents the result of importing a collection
type ImportResult struct {
	TotalCards    int      `json:"totalCards"`
	ImportedCards int      `json:"importedCards"`
	UpdatedCards  int      `json:"updatedCards"`
	Errors        []string `json:"errors"`
}
