package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/database"
)

type Deck struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Game        string    `json:"game"`
	IsPublic    bool      `json:"isPublic"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DeckCard struct {
	ID        uuid.UUID `json:"id"`
	DeckID    uuid.UUID `json:"deckId"`
	CardID    uuid.UUID `json:"cardId"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DeckStore struct {
	db *sql.DB
}

func NewDeckStore(db *sql.DB) *DeckStore {
	return &DeckStore{db: db}
}

func (s *DeckStore) Create(deck *Deck) error {
	query := `
		INSERT INTO decks (id, user_id, name, description, game, is_public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		deck.ID,
		deck.UserID,
		deck.Name,
		deck.Description,
		deck.Game,
		deck.IsPublic,
	).Scan(&deck.CreatedAt, &deck.UpdatedAt)

	return err
}

func (s *DeckStore) FindByID(id uuid.UUID) (*Deck, error) {
	deck := &Deck{}
	query := `
		SELECT id, user_id, name, description, game, is_public, created_at, updated_at
		FROM decks
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&deck.ID,
		&deck.UserID,
		&deck.Name,
		&deck.Description,
		&deck.Game,
		&deck.IsPublic,
		&deck.CreatedAt,
		&deck.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return deck, err
}

func (s *DeckStore) FindByUserID(userID uuid.UUID) ([]*Deck, error) {
	query := `
		SELECT id, user_id, name, description, game, is_public, created_at, updated_at
		FROM decks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decks []*Deck
	for rows.Next() {
		deck := &Deck{}
		err := rows.Scan(
			&deck.ID,
			&deck.UserID,
			&deck.Name,
			&deck.Description,
			&deck.Game,
			&deck.IsPublic,
			&deck.CreatedAt,
			&deck.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		decks = append(decks, deck)
	}

	return decks, nil
}

func (s *DeckStore) Update(deck *Deck) error {
	query := `
		UPDATE decks
		SET name = $1, description = $2, game = $3, is_public = $4
		WHERE id = $5
		RETURNING updated_at
	`

	return s.db.QueryRow(
		query,
		deck.Name,
		deck.Description,
		deck.Game,
		deck.IsPublic,
		deck.ID,
	).Scan(&deck.UpdatedAt)
}

func (s *DeckStore) Delete(id uuid.UUID) error {
	query := `DELETE FROM decks WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

func (s *DeckStore) AddCard(deckID, cardID uuid.UUID, quantity int) error {
	query := `
		INSERT INTO deck_cards (deck_id, card_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (deck_id, card_id) DO UPDATE
		SET quantity = deck_cards.quantity + $3
		RETURNING id, created_at, updated_at
	`

	var id uuid.UUID
	var createdAt, updatedAt time.Time
	return s.db.QueryRow(query, deckID, cardID, quantity).Scan(&id, &createdAt, &updatedAt)
}

func (s *DeckStore) RemoveCard(deckID, cardID uuid.UUID) error {
	query := `DELETE FROM deck_cards WHERE deck_id = $1 AND card_id = $2`
	_, err := s.db.Exec(query, deckID, cardID)
	return err
}

func (s *DeckStore) GetCards(deckID uuid.UUID) ([]*DeckCard, error) {
	query := `
		SELECT id, deck_id, card_id, quantity, created_at, updated_at
		FROM deck_cards
		WHERE deck_id = $1
	`

	rows, err := s.db.Query(query, deckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deckCards []*DeckCard
	for rows.Next() {
		deckCard := &DeckCard{}
		err := rows.Scan(
			&deckCard.ID,
			&deckCard.DeckID,
			&deckCard.CardID,
			&deckCard.Quantity,
			&deckCard.CreatedAt,
			&deckCard.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		deckCards = append(deckCards, deckCard)
	}

	return deckCards, nil
}

func (s *DeckStore) CreateWithCards(ctx context.Context, deck *Deck, cards []*DeckCard) error {
	return database.WithTransaction(ctx, s.db, func(tx *database.Transaction) error {
		// Create the deck
		query := `
			INSERT INTO decks (id, user_id, name, description, game, is_public)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING created_at, updated_at
		`

		err := tx.QueryRow(
			query,
			deck.ID,
			deck.UserID,
			deck.Name,
			deck.Description,
			deck.Game,
			deck.IsPublic,
		).Scan(&deck.CreatedAt, &deck.UpdatedAt)
		if err != nil {
			return err
		}

		// Add cards to the deck
		for _, card := range cards {
			query := `
				INSERT INTO deck_cards (deck_id, card_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING created_at, updated_at
			`

			err := tx.QueryRow(
				query,
				deck.ID,
				card.CardID,
				card.Quantity,
			).Scan(&card.CreatedAt, &card.UpdatedAt)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *DeckStore) UpdateWithCards(ctx context.Context, deck *Deck, cards []*DeckCard) error {
	return database.WithTransaction(ctx, s.db, func(tx *database.Transaction) error {
		// Update the deck
		query := `
			UPDATE decks
			SET name = $1, description = $2, game = $3, is_public = $4
			WHERE id = $5
			RETURNING updated_at
		`

		err := tx.QueryRow(
			query,
			deck.Name,
			deck.Description,
			deck.Game,
			deck.IsPublic,
			deck.ID,
		).Scan(&deck.UpdatedAt)
		if err != nil {
			return err
		}

		// Delete existing cards
		_, err = tx.Exec("DELETE FROM deck_cards WHERE deck_id = $1", deck.ID)
		if err != nil {
			return err
		}

		// Add new cards
		for _, card := range cards {
			query := `
				INSERT INTO deck_cards (deck_id, card_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING created_at, updated_at
			`

			err := tx.QueryRow(
				query,
				deck.ID,
				card.CardID,
				card.Quantity,
			).Scan(&card.CreatedAt, &card.UpdatedAt)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *DeckStore) DeleteWithCards(ctx context.Context, id uuid.UUID) error {
	return database.WithTransaction(ctx, s.db, func(tx *database.Transaction) error {
		// Delete the deck (this will cascade delete deck_cards due to foreign key constraint)
		_, err := tx.Exec("DELETE FROM decks WHERE id = $1", id)
		return err
	})
}

// GetDeckCard returns a deck card by its ID
func (s *DeckStore) GetDeckCard(id uuid.UUID) (*DeckCard, error) {
	deckCard := &DeckCard{}
	query := `
		SELECT id, deck_id, card_id, quantity, created_at, updated_at
		FROM deck_cards
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&deckCard.ID,
		&deckCard.DeckID,
		&deckCard.CardID,
		&deckCard.Quantity,
		&deckCard.CreatedAt,
		&deckCard.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return deckCard, err
}

// UpdateDeckCard updates the quantity of a card in a deck
func (s *DeckStore) UpdateDeckCard(id uuid.UUID, quantity int) error {
	query := `
		UPDATE deck_cards
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := s.db.Exec(query, quantity, id)
	return err
}
