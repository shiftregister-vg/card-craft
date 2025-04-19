package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver
type Resolver struct {
	authService     *auth.Service
	cardStore       *cards.CardStore
	deckStore       *models.DeckStore
	collectionStore *models.CollectionStore
	searchService   *cards.SearchService
}

// NewResolver creates a new resolver
func NewResolver(
	authService *auth.Service,
	cardStore *cards.CardStore,
	deckStore *models.DeckStore,
	collectionStore *models.CollectionStore,
	searchService *cards.SearchService,
) *Resolver {
	return &Resolver{
		authService:     authService,
		cardStore:       cardStore,
		deckStore:       deckStore,
		collectionStore: collectionStore,
		searchService:   searchService,
	}
}
