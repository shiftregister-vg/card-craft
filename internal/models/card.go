package models

import (
	"context"
	"database/sql"
	"fmt"
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
	ImageUrl  string    `json:"imageUrl"` // URL to the card image
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CardConnection represents a paginated connection of cards
type CardConnection struct {
	Edges    []*CardEdge `json:"edges"`
	PageInfo *PageInfo   `json:"pageInfo"`
}

// CardEdge represents an edge in a card connection
type CardEdge struct {
	Node   *Card  `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
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
		card.ImageUrl,
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
		&card.ImageUrl,
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
		WHERE LOWER(game) = LOWER($1)
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
			&card.ImageUrl,
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
			&card.ImageUrl,
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
		card.ImageUrl,
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

func (s *CardStore) FindByGameSetAndNumber(game string, setCode string, number string) (*Card, error) {
	var card Card

	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE game = $1 AND set_code = $2 AND number = $3
	`

	err := s.db.QueryRow(query, game, setCode, number).Scan(
		&card.ID,
		&card.Name,
		&card.Game,
		&card.SetCode,
		&card.SetName,
		&card.Number,
		&card.Rarity,
		&card.ImageUrl,
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

// SearchCards searches for cards based on the given criteria
func (s *CardStore) SearchCards(ctx context.Context, game, setCode, rarity, name *string, page, pageSize *int, sortBy, sortOrder *string) (*CardConnection, error) {
	// Build the query
	query := `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url,
			created_at, updated_at
		FROM cards
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if game != nil {
		query += fmt.Sprintf(" AND LOWER(game) = LOWER($%d)", argCount)
		args = append(args, *game)
		argCount++
	}

	if setCode != nil {
		query += fmt.Sprintf(" AND set_code = $%d", argCount)
		args = append(args, *setCode)
		argCount++
	}

	if rarity != nil {
		query += fmt.Sprintf(" AND rarity = $%d", argCount)
		args = append(args, *rarity)
		argCount++
	}

	if name != nil {
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+*name+"%")
		argCount++
	}

	// Add sorting
	if sortBy != nil {
		query += fmt.Sprintf(" ORDER BY %s", *sortBy)
		if sortOrder != nil {
			query += fmt.Sprintf(" %s", *sortOrder)
		}
	} else {
		query += " ORDER BY set_code, number"
	}

	// Add pagination
	if pageSize != nil {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, *pageSize)
		argCount++

		if page != nil {
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, (*page-1)*(*pageSize))
		}
	} else {
		query += " LIMIT 100"
	}

	// Execute the query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []*CardEdge
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
			&card.ImageUrl,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		edges = append(edges, &CardEdge{
			Node:   card,
			Cursor: fmt.Sprintf("%s:%s", card.SetCode, card.Number),
		})
	}

	// Check if there are more results
	var hasNextPage bool
	if len(edges) > 0 {
		checkQuery := `
			SELECT EXISTS (
				SELECT 1
				FROM cards
				WHERE 1=1
		`
		checkArgs := []interface{}{}
		checkArgCount := 1

		if game != nil {
			checkQuery += fmt.Sprintf(" AND LOWER(game) = LOWER($%d)", checkArgCount)
			checkArgs = append(checkArgs, *game)
			checkArgCount++
		}

		if setCode != nil {
			checkQuery += fmt.Sprintf(" AND set_code = $%d", checkArgCount)
			checkArgs = append(checkArgs, *setCode)
			checkArgCount++
		}

		if rarity != nil {
			checkQuery += fmt.Sprintf(" AND rarity = $%d", checkArgCount)
			checkArgs = append(checkArgs, *rarity)
			checkArgCount++
		}

		if name != nil {
			checkQuery += fmt.Sprintf(" AND name ILIKE $%d", checkArgCount)
			checkArgs = append(checkArgs, "%"+*name+"%")
			checkArgCount++
		}

		checkQuery += fmt.Sprintf(" AND (set_code > $%d OR (set_code = $%d AND number > $%d))", checkArgCount, checkArgCount, checkArgCount+1)
		checkArgs = append(checkArgs, edges[len(edges)-1].Node.SetCode, edges[len(edges)-1].Node.Number)

		checkQuery += " LIMIT 1)"

		err = s.db.QueryRowContext(ctx, checkQuery, checkArgs...).Scan(&hasNextPage)
		if err != nil {
			return nil, err
		}
	}

	// Generate the next cursor
	var endCursor *string
	if hasNextPage && len(edges) > 0 {
		cursor := fmt.Sprintf("%s:%s", edges[len(edges)-1].Node.SetCode, edges[len(edges)-1].Node.Number)
		endCursor = &cursor
	}

	return &CardConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   endCursor,
		},
	}, nil
}
