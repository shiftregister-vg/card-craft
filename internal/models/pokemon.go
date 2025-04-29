package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PokemonCard represents a Pokemon trading card
type PokemonCard struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	HP          int       `json:"hp"`
	Types       []string  `json:"types"`
	EvolvesFrom string    `json:"evolvesFrom"`
	EvolvesTo   []string  `json:"evolvesTo"`
	Abilities   []string  `json:"abilities"`
	Attacks     []string  `json:"attacks"`
	Set         string    `json:"set"`
	SetNumber   string    `json:"setNumber"`
	Rarity      string    `json:"rarity"`
	CardType    string    `json:"cardType"`
	Subtype     string    `json:"subtype"`
	Description string    `json:"description"`
	ImageUrl    string    `json:"imageUrl"`
	ReleasedAt  time.Time `json:"releasedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PokemonCardStore handles database operations for Pokemon cards
type PokemonCardStore struct {
	db *sql.DB
}

// NewPokemonCardStore creates a new Pokemon card store
func NewPokemonCardStore(db *sql.DB) *PokemonCardStore {
	return &PokemonCardStore{db: db}
}

// GetCardByCardID retrieves a Pokemon card by its card ID
func (s *PokemonCardStore) GetCardByCardID(ctx context.Context, cardID uuid.UUID) (*PokemonCard, error) {
	query := `
		SELECT id, name, hp, types, evolves_from, evolves_to, abilities, attacks,
			set_name, set_number, rarity, card_type, subtype, description, image_url,
			released_at, created_at, updated_at
		FROM pokemon_cards
		WHERE id = $1
	`

	var card PokemonCard
	err := s.db.QueryRowContext(ctx, query, cardID).Scan(
		&card.ID,
		&card.Name,
		&card.HP,
		&card.Types,
		&card.EvolvesFrom,
		&card.EvolvesTo,
		&card.Abilities,
		&card.Attacks,
		&card.Set,
		&card.SetNumber,
		&card.Rarity,
		&card.CardType,
		&card.Subtype,
		&card.Description,
		&card.ImageUrl,
		&card.ReleasedAt,
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}
