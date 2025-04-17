package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Card represents a trading card in the system
type Card struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Game      string    `json:"game"`     // e.g., "pokemon", "starwars", "lorcana"
	SetCode   string    `json:"setCode"`  // e.g., "SWSH01", "SWU01", "DRE"
	SetName   string    `json:"setName"`  // e.g., "Sword & Shield", "Spark of Rebellion", "The First Chapter"
	Number    string    `json:"number"`   // e.g., "001/264", "001"
	Rarity    string    `json:"rarity"`   // e.g., "Common", "Uncommon", "Rare", "Holo Rare"
	ImageURL  string    `json:"imageUrl"` // URL to the card image
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CardStore handles database operations for cards
type CardStore struct {
	db *sql.DB
}

// NewCardStore creates a new CardStore
func NewCardStore(db *sql.DB) *CardStore {
	return &CardStore{db: db}
}

// Create inserts a new card into the database
func (s *CardStore) Create(card *Card) error {
	query := `
		INSERT INTO cards (
			id, name, game, set_code, set_name, number, rarity, image_url,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	_, err := s.db.Exec(
		query,
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

// FindByID retrieves a card by its ID
func (s *CardStore) FindByID(id uuid.UUID) (*Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url,
			created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &Card{}
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

	return card, err
}

// FindByGame retrieves all cards for a specific game
func (s *CardStore) FindByGame(game string) ([]*Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url,
			created_at, updated_at
		FROM cards
		WHERE game = $1
		ORDER BY set_code, number
	`

	rows, err := s.db.Query(query, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card := &Card{}
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

	return cards, rows.Err()
}

// FindBySet retrieves all cards in a specific set
func (s *CardStore) FindBySet(game, setCode string) ([]*Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url,
			created_at, updated_at
		FROM cards
		WHERE game = $1 AND set_code = $2
		ORDER BY number
	`

	rows, err := s.db.Query(query, game, setCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card := &Card{}
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

	return cards, rows.Err()
}

// Update updates an existing card
func (s *CardStore) Update(card *Card) error {
	query := `
		UPDATE cards
		SET name = $1, game = $2, set_code = $3, set_name = $4, number = $5,
			rarity = $6, image_url = $7, updated_at = $8
		WHERE id = $9
	`

	card.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		query,
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

// Delete removes a card from the database
func (s *CardStore) Delete(id uuid.UUID) error {
	query := `DELETE FROM cards WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
