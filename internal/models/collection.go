package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Collection represents a user's card collection
type Collection struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Game        string            `json:"game"`
	UserID      uuid.UUID         `json:"userId"`
	Cards       []*CollectionCard `json:"cards"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

// CollectionCard represents a card in a collection with its quantity and condition
type CollectionCard struct {
	ID           uuid.UUID `json:"id"`
	CollectionID uuid.UUID `json:"collectionId"`
	CardID       uuid.UUID `json:"cardId"`
	Card         *Card     `json:"card"`
	Quantity     int       `json:"quantity"`
	Condition    string    `json:"condition"`
	IsFoil       bool      `json:"isFoil"`
	Notes        string    `json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ImportError represents an error encountered during bulk import
type ImportError struct {
	CardID  string `json:"cardId"`
	Message string `json:"message"`
}

// BulkImportResult represents the result of a bulk import operation
type BulkImportResult struct {
	Success       bool           `json:"success"`
	ImportedCount int            `json:"importedCount"`
	Errors        []*ImportError `json:"errors,omitempty"`
}

// CollectionStore handles database operations for collections
type CollectionStore struct {
	db *sql.DB
}

// NewCollectionStore creates a new CollectionStore
func NewCollectionStore(db *sql.DB) *CollectionStore {
	return &CollectionStore{db: db}
}

// Create inserts a new collection into the database
func (s *CollectionStore) Create(collection *Collection) error {
	query := `
		INSERT INTO collections (
			id, user_id, name, description, game,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	collection.CreatedAt = now
	collection.UpdatedAt = now

	_, err := s.db.Exec(
		query,
		collection.ID,
		collection.UserID,
		collection.Name,
		collection.Description,
		collection.Game,
		collection.CreatedAt,
		collection.UpdatedAt,
	)

	return err
}

// FindByID retrieves a collection by its ID
func (s *CollectionStore) FindByID(id uuid.UUID) (*Collection, error) {
	query := `
		SELECT id, user_id, name, description, game,
			created_at, updated_at
		FROM collections
		WHERE id = $1
	`

	collection := &Collection{}
	err := s.db.QueryRow(query, id).Scan(
		&collection.ID,
		&collection.UserID,
		&collection.Name,
		&collection.Description,
		&collection.Game,
		&collection.CreatedAt,
		&collection.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load the collection cards
	cards, err := s.GetCards(id)
	if err != nil {
		return nil, err
	}
	collection.Cards = cards

	return collection, nil
}

// FindByUserID retrieves all collections for a user
func (s *CollectionStore) FindByUserID(userID uuid.UUID) ([]*Collection, error) {
	query := `
		SELECT id, user_id, name, description, game,
			created_at, updated_at
		FROM collections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []*Collection
	for rows.Next() {
		collection := &Collection{}
		err := rows.Scan(
			&collection.ID,
			&collection.UserID,
			&collection.Name,
			&collection.Description,
			&collection.Game,
			&collection.CreatedAt,
			&collection.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		collections = append(collections, collection)
	}

	return collections, rows.Err()
}

// Update updates an existing collection
func (s *CollectionStore) Update(collection *Collection) error {
	query := `
		UPDATE collections
		SET name = $1, description = $2, game = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`

	collection.UpdatedAt = time.Now()
	result, err := s.db.Exec(
		query,
		collection.Name,
		collection.Description,
		collection.Game,
		collection.UpdatedAt,
		collection.ID,
		collection.UserID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a collection from the database
func (s *CollectionStore) Delete(id, userID uuid.UUID) error {
	query := `DELETE FROM collections WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// AddCard adds a card to a collection
func (s *CollectionStore) AddCard(card *CollectionCard) error {
	query := `
		INSERT INTO collection_cards (
			id, collection_id, card_id, quantity, condition, is_foil, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	now := time.Now()
	card.CreatedAt = now
	card.UpdatedAt = now

	_, err := s.db.Exec(
		query,
		card.ID,
		card.CollectionID,
		card.CardID,
		card.Quantity,
		card.Condition,
		card.IsFoil,
		card.Notes,
		card.CreatedAt,
		card.UpdatedAt,
	)

	return err
}

// UpdateCard updates a card in a collection
func (s *CollectionStore) UpdateCard(card *CollectionCard) error {
	query := `
		UPDATE collection_cards
		SET quantity = $1, condition = $2, is_foil = $3, notes = $4, updated_at = $5
		WHERE id = $6
	`

	card.UpdatedAt = time.Now()
	result, err := s.db.Exec(
		query,
		card.Quantity,
		card.Condition,
		card.IsFoil,
		card.Notes,
		card.UpdatedAt,
		card.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// RemoveCard removes a card from a collection
func (s *CollectionStore) RemoveCard(id uuid.UUID) error {
	query := `DELETE FROM collection_cards WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetCards retrieves all cards in a collection
func (s *CollectionStore) GetCards(collectionID uuid.UUID) ([]*CollectionCard, error) {
	query := `
		SELECT 
			cc.id, cc.collection_id, cc.card_id, cc.quantity, cc.condition, cc.is_foil, cc.notes,
			cc.created_at, cc.updated_at,
			c.id, c.name, c.game, c.set_code, c.set_name, c.number, c.rarity, c.image_url,
			c.created_at, c.updated_at
		FROM collection_cards cc
		JOIN cards c ON cc.card_id = c.id
		WHERE cc.collection_id = $1
		ORDER BY cc.created_at DESC
	`

	rows, err := s.db.Query(query, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*CollectionCard
	for rows.Next() {
		var (
			condition *string
			notes     *string
		)
		card := &CollectionCard{
			Card: &Card{},
		}
		err := rows.Scan(
			&card.ID,
			&card.CollectionID,
			&card.CardID,
			&card.Quantity,
			&condition,
			&card.IsFoil,
			&notes,
			&card.CreatedAt,
			&card.UpdatedAt,
			&card.Card.ID,
			&card.Card.Name,
			&card.Card.Game,
			&card.Card.SetCode,
			&card.Card.SetName,
			&card.Card.Number,
			&card.Card.Rarity,
			&card.Card.ImageUrl,
			&card.Card.CreatedAt,
			&card.Card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if condition != nil {
			card.Condition = *condition
		}
		if notes != nil {
			card.Notes = *notes
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

// GetCard retrieves a specific card from a collection
func (s *CollectionStore) GetCard(id uuid.UUID) (*CollectionCard, error) {
	query := `
		SELECT 
			cc.id, cc.collection_id, cc.card_id, cc.quantity, cc.condition, cc.is_foil, cc.notes,
			cc.created_at, cc.updated_at,
			c.id, c.name, c.game, c.set_code, c.set_name, c.number, c.rarity, c.image_url,
			c.created_at, c.updated_at
		FROM collection_cards cc
		JOIN cards c ON cc.card_id = c.id
		WHERE cc.id = $1
	`

	var (
		condition string
		notes     string
	)

	card := &CollectionCard{
		Card:      &Card{},
		Condition: condition,
		Notes:     notes,
	}
	err := s.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.CollectionID,
		&card.CardID,
		&card.Quantity,
		&condition,
		&card.IsFoil,
		&notes,
		&card.CreatedAt,
		&card.UpdatedAt,
		&card.Card.ID,
		&card.Card.Name,
		&card.Card.Game,
		&card.Card.SetCode,
		&card.Card.SetName,
		&card.Card.Number,
		&card.Card.Rarity,
		&card.Card.ImageUrl,
		&card.Card.CreatedAt,
		&card.Card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	card.Condition = condition
	card.Notes = notes

	return card, err
}

// BulkAddCards adds multiple cards to a collection in a single transaction
func (s *CollectionStore) BulkAddCards(collectionID uuid.UUID, cards []*CollectionCardInput) (*BulkImportResult, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var importedCount int
	var errors []*ImportError

	query := `
		INSERT INTO collection_cards (
			id, collection_id, card_id, quantity, condition,
			is_foil, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (collection_id, card_id) DO UPDATE
		SET quantity = collection_cards.quantity + EXCLUDED.quantity,
			condition = COALESCE(EXCLUDED.condition, collection_cards.condition),
			is_foil = COALESCE(EXCLUDED.is_foil, collection_cards.is_foil),
			notes = COALESCE(EXCLUDED.notes, collection_cards.notes),
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now()

	for _, input := range cards {
		cardID, err := uuid.Parse(input.CardID)
		if err != nil {
			errors = append(errors, &ImportError{
				CardID:  input.CardID,
				Message: "invalid card ID",
			})
			continue
		}

		// Handle nullable fields
		var condition string
		if input.Condition != nil {
			condition = *input.Condition
		}

		var isFoil bool
		if input.IsFoil != nil {
			isFoil = *input.IsFoil
		}

		var notes string
		if input.Notes != nil {
			notes = *input.Notes
		}

		collectionCard := &CollectionCard{
			ID:           uuid.New(),
			CollectionID: collectionID,
			CardID:       cardID,
			Quantity:     input.Quantity,
			Condition:    condition,
			IsFoil:       isFoil,
			Notes:        notes,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = tx.Exec(
			query,
			collectionCard.ID,
			collectionCard.CollectionID,
			collectionCard.CardID,
			collectionCard.Quantity,
			collectionCard.Condition,
			collectionCard.IsFoil,
			collectionCard.Notes,
			collectionCard.CreatedAt,
			collectionCard.UpdatedAt,
		)

		if err != nil {
			errors = append(errors, &ImportError{
				CardID:  input.CardID,
				Message: err.Error(),
			})
			continue
		}

		importedCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &BulkImportResult{
		Success:       true,
		ImportedCount: importedCount,
		Errors:        errors,
	}, nil
}

// GetCollectionCard retrieves a collection card by ID
func (s *CollectionStore) GetCollectionCard(ctx context.Context, id string) (*CollectionCard, error) {
	var card CollectionCard
	err := s.db.QueryRowContext(ctx, `
		SELECT id, collection_id, card_id, quantity, condition, is_foil, notes, created_at, updated_at
		FROM collection_cards
		WHERE id = $1
	`, id).Scan(
		&card.ID,
		&card.CollectionID,
		&card.CardID,
		&card.Quantity,
		&card.Condition,
		&card.IsFoil,
		&card.Notes,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}
