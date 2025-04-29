package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// MTGCard represents a Magic: The Gathering card
type MTGCard struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	ManaCost      string    `json:"manaCost"`
	CMC           float64   `json:"cmc"`
	TypeLine      string    `json:"typeLine"`
	OracleText    string    `json:"oracleText"`
	Power         string    `json:"power"`
	Toughness     string    `json:"toughness"`
	Loyalty       string    `json:"loyalty"`
	Colors        []string  `json:"colors"`
	ColorIdentity []string  `json:"colorIdentity"`
	Keywords      []string  `json:"keywords"`
	Legalities    []string  `json:"legalities"`
	Reserved      bool      `json:"reserved"`
	Foil          bool      `json:"foil"`
	Nonfoil       bool      `json:"nonfoil"`
	Promo         bool      `json:"promo"`
	Reprint       bool      `json:"reprint"`
	Variation     bool      `json:"variation"`
	SetType       string    `json:"setType"`
	ReleasedAt    time.Time `json:"releasedAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// MTGCardStore handles database operations for MTG cards
type MTGCardStore struct {
	db *sql.DB
}

// NewMTGCardStore creates a new MTG card store
func NewMTGCardStore(db *sql.DB) *MTGCardStore {
	return &MTGCardStore{db: db}
}

// GetCardByCardID retrieves an MTG card by its card ID
func (s *MTGCardStore) GetCardByCardID(ctx context.Context, cardID uuid.UUID) (*MTGCard, error) {
	query := `
		SELECT id, name, mana_cost, cmc, type_line, oracle_text, power, toughness, loyalty,
			colors, color_identity, keywords, legalities, reserved, foil, nonfoil, promo,
			reprint, variation, set_type, released_at, created_at, updated_at
		FROM mtg_cards
		WHERE id = $1
	`

	var card MTGCard
	err := s.db.QueryRowContext(ctx, query, cardID).Scan(
		&card.ID,
		&card.Name,
		&card.ManaCost,
		&card.CMC,
		&card.TypeLine,
		&card.OracleText,
		&card.Power,
		&card.Toughness,
		&card.Loyalty,
		&card.Colors,
		&card.ColorIdentity,
		&card.Keywords,
		&card.Legalities,
		&card.Reserved,
		&card.Foil,
		&card.Nonfoil,
		&card.Promo,
		&card.Reprint,
		&card.Variation,
		&card.SetType,
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
