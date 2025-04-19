package types

import (
	"time"

	"github.com/google/uuid"
)

// Card represents a trading card in the database
type Card struct {
	ID        uuid.UUID
	Name      string
	Game      string
	SetCode   string
	SetName   string
	Number    string
	Rarity    string
	ImageURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

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

// DeckInput represents the input for creating or updating a deck
type DeckInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Game        string  `json:"game"`
}

// DeckCardInput represents the input for adding a card to a deck
type DeckCardInput struct {
	CardID   string `json:"cardId"`
	Quantity int    `json:"quantity"`
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
	Cards    []*Card `json:"cards"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"pageSize"`
}
