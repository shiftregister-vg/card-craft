package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"database/sql"

	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/shiftregister-vg/card-craft/internal/search"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver serves as dependency injection for your app, add any dependencies you require here.
type Resolver struct {
	db              *sql.DB
	cardStore       *cards.CardStore
	pokemonStore    *cards.PokemonCardStore
	deckStore       *models.DeckStore
	collectionStore *models.CollectionStore
	authService     *auth.Service
	searchService   *search.SearchService
}

// NewResolver creates a new resolver with the given dependencies
func NewResolver(
	db *sql.DB,
	cardStore *cards.CardStore,
	pokemonStore *cards.PokemonCardStore,
	deckStore *models.DeckStore,
	collectionStore *models.CollectionStore,
	authService *auth.Service,
) *Resolver {
	searchService := search.NewSearchService(cardStore)
	return &Resolver{
		db:              db,
		cardStore:       cardStore,
		pokemonStore:    pokemonStore,
		deckStore:       deckStore,
		collectionStore: collectionStore,
		authService:     authService,
		searchService:   searchService,
	}
}
