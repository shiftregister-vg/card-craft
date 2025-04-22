package cards

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// CardStore handles database operations for cards
type CardStore struct {
	db *sql.DB
}

// NewCardStore creates a new card store
func NewCardStore(db *sql.DB) *CardStore {
	return &CardStore{db: db}
}

// Create inserts a new card into the database
func (s *CardStore) Create(card *types.Card) error {
	query := `
		INSERT INTO cards (id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.Exec(query,
		card.ID,
		card.Name,
		card.Game,
		card.SetCode,
		card.SetName,
		card.Number,
		card.Rarity,
		card.ImageURL,
		card.CreatedAt,
		card.UpdatedAt,
	)
	return err
}

// Update updates an existing card in the database
func (s *CardStore) Update(card *types.Card) error {
	query := `
		UPDATE cards
		SET name = $1, game = $2, set_code = $3, set_name = $4, number = $5, rarity = $6, image_url = $7, updated_at = $8
		WHERE id = $9
	`
	_, err := s.db.Exec(query,
		card.Name,
		card.Game,
		card.SetCode,
		card.SetName,
		card.Number,
		card.Rarity,
		card.ImageURL,
		card.UpdatedAt,
		card.ID,
	)
	return err
}

// FindByGameAndNumber finds a card by its game, set code, and number
func (s *CardStore) FindByGameAndNumber(game, setCode, number string) (*types.Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE game = $1 AND set_code = $2 AND number = $3
	`
	row := s.db.QueryRow(query, game, setCode, number)

	var card types.Card
	err := row.Scan(
		&card.ID,
		&card.Name,
		&card.Game,
		&card.SetCode,
		&card.SetName,
		&card.Number,
		&card.Rarity,
		&card.ImageURL,
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

// SearchCards searches for cards by name and game
func (s *CardStore) SearchCards(query string, game string) ([]*types.Card, error) {
	sqlQuery := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE game = $1 AND name ILIKE $2
		ORDER BY name
		LIMIT 50
	`
	rows, err := s.db.Query(sqlQuery, game, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*types.Card
	for rows.Next() {
		var card types.Card
		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.Game,
			&card.SetCode,
			&card.SetName,
			&card.Number,
			&card.Rarity,
			&card.ImageURL,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}
	return cards, nil
}

func (s *CardStore) FindByID(id uuid.UUID) (*types.Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &types.Card{}
	err := s.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.Name,
		&card.Game,
		&card.SetCode,
		&card.SetName,
		&card.Number,
		&card.Rarity,
		&card.ImageURL,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (s *CardStore) FindByGame(game string, first int, after string) ([]*types.Card, string, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE LOWER(game) = LOWER($1)
		AND ($2 = '' OR (set_code > $2) OR (set_code = $2 AND number > $3))
		ORDER BY set_code, number
		LIMIT $4
	`

	// Parse the after cursor to get set_code and number
	var setCode, number string
	if after != "" {
		parts := strings.Split(after, ":")
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid cursor format")
		}
		setCode, number = parts[0], parts[1]
	}

	rows, err := s.db.Query(query, game, setCode, number, first)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var cards []*types.Card
	var lastSetCode, lastNumber string
	for rows.Next() {
		card := &types.Card{}
		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.Game,
			&card.SetCode,
			&card.SetName,
			&card.Number,
			&card.Rarity,
			&card.ImageURL,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, "", err
		}
		cards = append(cards, card)
		lastSetCode = card.SetCode
		lastNumber = card.Number
	}

	// Check if there are more results
	var hasMore bool
	if len(cards) > 0 {
		checkQuery := `
			SELECT EXISTS (
				SELECT 1
				FROM cards
				WHERE LOWER(game) = LOWER($1)
				AND (set_code > $2 OR (set_code = $2 AND number > $3))
				LIMIT 1
			)
		`
		err = s.db.QueryRow(checkQuery, game, lastSetCode, lastNumber).Scan(&hasMore)
		if err != nil {
			return nil, "", err
		}
	}

	// Generate the next cursor
	var nextCursor string
	if hasMore && len(cards) > 0 {
		nextCursor = fmt.Sprintf("%s:%s", lastSetCode, lastNumber)
	}

	return cards, nextCursor, nil
}

func (s *CardStore) FindBySet(game, setCode string) ([]*types.Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE game = $1 AND set_code = $2
		ORDER BY number
	`

	rows, err := s.db.Query(query, game, setCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*types.Card
	for rows.Next() {
		card := &types.Card{}
		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.Game,
			&card.SetCode,
			&card.SetName,
			&card.Number,
			&card.Rarity,
			&card.ImageURL,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (s *CardStore) Delete(id uuid.UUID) error {
	query := `DELETE FROM cards WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// ToModel converts a types.Card to a models.Card
func (s *CardStore) ToModel(card *types.Card) *models.Card {
	return &models.Card{
		ID:        card.ID,
		Name:      card.Name,
		Game:      card.Game,
		SetCode:   card.SetCode,
		SetName:   card.SetName,
		Number:    card.Number,
		Rarity:    card.Rarity,
		ImageUrl:  card.ImageURL,
		CreatedAt: card.CreatedAt,
		UpdatedAt: card.UpdatedAt,
	}
}
