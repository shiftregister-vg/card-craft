package cards

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(card *models.Card) error {
	query := `
		INSERT INTO cards (
			id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
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
		now,
		now,
	)
	return err
}

func (s *Store) FindByID(id uuid.UUID) (*models.Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE id = $1
	`

	card := &models.Card{}
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

func (s *Store) FindByGame(game string) ([]*models.Card, error) {
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE game = $1
		ORDER BY set_code, number
	`

	rows, err := s.db.Query(query, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
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

func (s *Store) FindBySet(game, setCode string) ([]*models.Card, error) {
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

	var cards []*models.Card
	for rows.Next() {
		card := &models.Card{}
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

func (s *Store) Update(card *models.Card) error {
	query := `
		UPDATE cards
		SET name = $1, game = $2, set_code = $3, set_name = $4, number = $5, rarity = $6, image_url = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := s.db.Exec(
		query,
		card.Name,
		card.Game,
		card.SetCode,
		card.SetName,
		card.Number,
		card.Rarity,
		card.ImageURL,
		time.Now(),
		card.ID,
	)
	return err
}

func (s *Store) Delete(id uuid.UUID) error {
	query := `DELETE FROM cards WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
