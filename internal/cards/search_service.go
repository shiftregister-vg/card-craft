package cards

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// SearchOptions represents the options for card search
type SearchOptions struct {
	Game      string
	SetCode   string
	Rarity    string
	Name      string
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// SearchResult represents the result of a card search
type SearchResult struct {
	Cards      []*models.Card
	TotalCount int
	Page       int
	PageSize   int
}

// SearchService handles card search and filtering
type SearchService struct {
	db *sql.DB
}

// NewSearchService creates a new SearchService
func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// Search searches for cards based on the provided options
func (s *SearchService) Search(opts types.SearchOptions) (*types.CardSearchResult, error) {
	// Build the base query
	query := "SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at FROM cards WHERE 1=1"
	args := []interface{}{}

	// Add filters
	if opts.Game != "" {
		query += " AND game = ?"
		args = append(args, opts.Game)
	}
	if opts.SetCode != "" {
		query += " AND set_code = ?"
		args = append(args, opts.SetCode)
	}
	if opts.Rarity != "" {
		query += " AND rarity = ?"
		args = append(args, opts.Rarity)
	}
	if opts.Name != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+opts.Name+"%")
	}

	// Add sorting
	if opts.SortBy != "" {
		query += " ORDER BY " + opts.SortBy
		if opts.SortOrder != "" {
			query += " " + opts.SortOrder
		}
	}

	// Add pagination
	if opts.PageSize > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, opts.PageSize, (opts.Page-1)*opts.PageSize)
	}

	// Execute the query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}
	defer rows.Close()

	// Process results
	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		var createdAt, updatedAt time.Time
		err := rows.Scan(
			&card.ID,
			&card.Name,
			&card.Game,
			&card.SetCode,
			&card.SetName,
			&card.Number,
			&card.Rarity,
			&card.ImageURL,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		card.CreatedAt = createdAt
		card.UpdatedAt = updatedAt
		cards = append(cards, &card)
	}

	// Get total count
	var total int
	countQuery := strings.Split(query, "ORDER BY")[0]
	countQuery = strings.Replace(countQuery, "SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at", "SELECT COUNT(*)", 1)
	countQuery = strings.Replace(countQuery, "LIMIT ? OFFSET ?", "", 1)
	err = s.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	return &types.CardSearchResult{
		Cards:    cards,
		Total:    total,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GetFilters returns the available filters for a game
func (s *SearchService) GetFilters(game string) (*types.CardFilters, error) {
	filters := &types.CardFilters{}

	// Get distinct set codes
	rows, err := s.db.Query("SELECT DISTINCT set_code FROM cards WHERE game = ? ORDER BY set_code", game)
	if err != nil {
		return nil, fmt.Errorf("failed to get set codes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var set string
		if err := rows.Scan(&set); err != nil {
			return nil, fmt.Errorf("failed to scan set code: %w", err)
		}
		filters.Sets = append(filters.Sets, set)
	}

	// Get distinct rarities
	rows, err = s.db.Query("SELECT DISTINCT rarity FROM cards WHERE game = ? ORDER BY rarity", game)
	if err != nil {
		return nil, fmt.Errorf("failed to get rarities: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rarity string
		if err := rows.Scan(&rarity); err != nil {
			return nil, fmt.Errorf("failed to scan rarity: %w", err)
		}
		filters.Rarities = append(filters.Rarities, rarity)
	}

	return filters, nil
}
