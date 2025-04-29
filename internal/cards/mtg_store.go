package cards

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// MTGCard represents MTG-specific card data
type MTGCard struct {
	ID            string
	CardID        string
	ManaCost      string
	CMC           float64
	TypeLine      string
	OracleText    string
	Power         string
	Toughness     string
	Loyalty       string
	Colors        []string
	ColorIdentity []string
	Keywords      []string
	Legalities    map[string]string
	Reserved      bool
	Foil          bool
	Nonfoil       bool
	Promo         bool
	Reprint       bool
	Variation     bool
	SetType       string
	ReleasedAt    time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MTGCardStore handles database operations for MTG-specific card data
type MTGCardStore struct {
	db *sql.DB
}

// NewMTGCardStore creates a new MTG card store
func NewMTGCardStore(db *sql.DB) *MTGCardStore {
	return &MTGCardStore{db: db}
}

// Create inserts a new MTG card into the database
func (s *MTGCardStore) Create(ctx context.Context, card *MTGCard) error {
	legalitiesJSON, err := json.Marshal(card.Legalities)
	if err != nil {
		return fmt.Errorf("failed to marshal legalities: %w", err)
	}

	query := `
		INSERT INTO mtg_cards (
			card_id, mana_cost, cmc, type_line, oracle_text, power, toughness, loyalty,
			colors, color_identity, keywords, legalities, reserved, foil, nonfoil,
			promo, reprint, variation, set_type, released_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)
		RETURNING id, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		card.CardID, card.ManaCost, card.CMC, card.TypeLine, card.OracleText,
		card.Power, card.Toughness, card.Loyalty,
		pq.Array(card.Colors), pq.Array(card.ColorIdentity), pq.Array(card.Keywords),
		legalitiesJSON, card.Reserved, card.Foil, card.Nonfoil,
		card.Promo, card.Reprint, card.Variation, card.SetType, card.ReleasedAt,
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create MTG card: %w", err)
	}

	return nil
}

// Update updates an existing MTG card in the database
func (s *MTGCardStore) Update(ctx context.Context, card *MTGCard) error {
	legalitiesJSON, err := json.Marshal(card.Legalities)
	if err != nil {
		return fmt.Errorf("failed to marshal legalities: %w", err)
	}

	query := `
		UPDATE mtg_cards SET
			mana_cost = $1, cmc = $2, type_line = $3, oracle_text = $4,
			power = $5, toughness = $6, loyalty = $7, colors = $8,
			color_identity = $9, keywords = $10, legalities = $11,
			reserved = $12, foil = $13, nonfoil = $14, promo = $15,
			reprint = $16, variation = $17, set_type = $18, released_at = $19
		WHERE card_id = $20
		RETURNING updated_at
	`

	err = s.db.QueryRowContext(ctx, query,
		card.ManaCost, card.CMC, card.TypeLine, card.OracleText,
		card.Power, card.Toughness, card.Loyalty,
		pq.Array(card.Colors), pq.Array(card.ColorIdentity), pq.Array(card.Keywords),
		legalitiesJSON, card.Reserved, card.Foil, card.Nonfoil,
		card.Promo, card.Reprint, card.Variation, card.SetType, card.ReleasedAt,
		card.CardID,
	).Scan(&card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update MTG card: %w", err)
	}

	return nil
}

// FindByCardID finds an MTG card by its card ID
func (s *MTGCardStore) FindByCardID(ctx context.Context, cardID string) (*MTGCard, error) {
	query := `
		SELECT
			id, card_id, mana_cost, cmc, type_line, oracle_text,
			power, toughness, loyalty, colors, color_identity,
			keywords, legalities, reserved, foil, nonfoil,
			promo, reprint, variation, set_type, released_at,
			created_at, updated_at
		FROM mtg_cards
		WHERE card_id = $1
	`

	var card MTGCard
	var legalitiesJSON []byte

	err := s.db.QueryRowContext(ctx, query, cardID).Scan(
		&card.ID, &card.CardID, &card.ManaCost, &card.CMC, &card.TypeLine,
		&card.OracleText, &card.Power, &card.Toughness, &card.Loyalty,
		pq.Array(&card.Colors), pq.Array(&card.ColorIdentity), pq.Array(&card.Keywords),
		&legalitiesJSON, &card.Reserved, &card.Foil, &card.Nonfoil,
		&card.Promo, &card.Reprint, &card.Variation, &card.SetType, &card.ReleasedAt,
		&card.CreatedAt, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find MTG card: %w", err)
	}

	if err := json.Unmarshal(legalitiesJSON, &card.Legalities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal legalities: %w", err)
	}

	return &card, nil
}

// Delete removes an MTG card from the database
func (s *MTGCardStore) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM mtg_cards WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete MTG card: %w", err)
	}
	return nil
}

// CreateBatch inserts multiple MTG cards into the database
func (s *MTGCardStore) CreateBatch(ctx context.Context, tx *sql.Tx, cards []*MTGCard) error {
	query := `
		INSERT INTO mtg_cards (
			card_id, mana_cost, cmc, type_line, oracle_text, power, toughness, loyalty,
			colors, color_identity, keywords, legalities, reserved, foil, nonfoil,
			promo, reprint, variation, set_type, released_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22
		)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, card := range cards {
		legalitiesJSON, err := json.Marshal(card.Legalities)
		if err != nil {
			return fmt.Errorf("failed to marshal legalities: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			card.CardID,
			card.ManaCost,
			card.CMC,
			card.TypeLine,
			card.OracleText,
			card.Power,
			card.Toughness,
			card.Loyalty,
			pq.Array(card.Colors),
			pq.Array(card.ColorIdentity),
			pq.Array(card.Keywords),
			legalitiesJSON,
			card.Reserved,
			card.Foil,
			card.Nonfoil,
			card.Promo,
			card.Reprint,
			card.Variation,
			card.SetType,
			card.ReleasedAt,
			card.CreatedAt,
			card.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to create MTG card: %w", err)
		}
	}
	return nil
}

// UpdateBatch updates multiple MTG cards in the database
func (s *MTGCardStore) UpdateBatch(ctx context.Context, tx *sql.Tx, cards []*MTGCard) error {
	query := `
		UPDATE mtg_cards
		SET mana_cost = $1, cmc = $2, type_line = $3, oracle_text = $4, power = $5,
			toughness = $6, loyalty = $7, colors = $8, color_identity = $9, keywords = $10,
			legalities = $11, reserved = $12, foil = $13, nonfoil = $14, promo = $15,
			reprint = $16, variation = $17, set_type = $18, released_at = $19, updated_at = $20
		WHERE id = $21
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, card := range cards {
		legalitiesJSON, err := json.Marshal(card.Legalities)
		if err != nil {
			return fmt.Errorf("failed to marshal legalities: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			card.ManaCost,
			card.CMC,
			card.TypeLine,
			card.OracleText,
			card.Power,
			card.Toughness,
			card.Loyalty,
			pq.Array(card.Colors),
			pq.Array(card.ColorIdentity),
			pq.Array(card.Keywords),
			legalitiesJSON,
			card.Reserved,
			card.Foil,
			card.Nonfoil,
			card.Promo,
			card.Reprint,
			card.Variation,
			card.SetType,
			card.ReleasedAt,
			card.UpdatedAt,
			card.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update MTG card: %w", err)
		}
	}
	return nil
}
