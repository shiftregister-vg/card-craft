package cards

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// CardImporter defines the interface for importing cards
type CardImporter interface {
	Import(ctx context.Context, store *CardStore) error
	GetGame() string
}

// Importer defines the interface for card importers
type Importer interface {
	Import(ctx context.Context, store *CardStore) error
}

// BaseImporter provides common functionality for all card importers
type BaseImporter struct {
	store *CardStore
	game  string
}

// NewBaseImporter creates a new base importer
func NewBaseImporter(store *CardStore, game string) *BaseImporter {
	return &BaseImporter{
		store: store,
		game:  game,
	}
}

// GetGame returns the game identifier
func (i *BaseImporter) GetGame() string {
	return i.game
}

// CreateOrUpdateCard handles the common logic for creating or updating a card
func (i *BaseImporter) CreateOrUpdateCard(card *types.Card) error {
	existing, err := i.store.FindByGameAndNumber(i.game, card.SetCode, card.Number)
	if err != nil {
		return err
	}

	if existing != nil {
		card.ID = existing.ID
		card.CreatedAt = existing.CreatedAt
		return i.store.Update(card)
	}

	return i.store.Create(card)
}

// NewCard creates a new card with common fields set
func (i *BaseImporter) NewCard() *types.Card {
	now := time.Now()
	return &types.Card{
		ID:        uuid.New(),
		Game:      i.game,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
