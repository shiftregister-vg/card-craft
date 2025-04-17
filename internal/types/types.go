package types

import (
	"github.com/shiftregister-vg/card-craft/internal/models"
)

// CardInput represents the input for creating or updating a card
type CardInput struct {
	Name     string  `json:"name"`
	Game     string  `json:"game"`
	SetCode  string  `json:"setCode"`
	SetName  string  `json:"setName"`
	Number   string  `json:"number"`
	Rarity   string  `json:"rarity"`
	ImageURL *string `json:"imageUrl"`
}

// SearchOptions represents the options for searching cards
type SearchOptions struct {
	Game      string `json:"game"`
	SetCode   string `json:"setCode"`
	Rarity    string `json:"rarity"`
	Name      string `json:"name"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	SortBy    string `json:"sortBy"`
	SortOrder string `json:"sortOrder"`
}

// CardFilters represents the available filters for a game
type CardFilters struct {
	Sets     []string `json:"sets"`
	Rarities []string `json:"rarities"`
}

// CardSearchResult represents the result of a card search
type CardSearchResult struct {
	Cards    []*models.Card `json:"cards"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}
