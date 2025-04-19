package cards

import (
	"fmt"
	"strings"

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
	Cards      []*types.Card
	TotalCount int
	Page       int
	PageSize   int
}

// SearchService handles card search and filtering
type SearchService struct {
	cardStore *CardStore
}

// NewSearchService creates a new SearchService
func NewSearchService(cardStore *CardStore) *SearchService {
	return &SearchService{cardStore: cardStore}
}

// Search searches for cards based on the provided options
func (s *SearchService) Search(opts types.SearchOptions) (*types.CardSearchResult, error) {
	// Get all cards for the game
	cards, err := s.cardStore.FindByGame(opts.Game)
	if err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}

	// Apply filters
	var filteredCards []*types.Card
	for _, card := range cards {
		if opts.SetCode != "" && card.SetCode != opts.SetCode {
			continue
		}
		if opts.Rarity != "" && card.Rarity != opts.Rarity {
			continue
		}
		if opts.Name != "" && !strings.Contains(strings.ToLower(card.Name), strings.ToLower(opts.Name)) {
			continue
		}
		filteredCards = append(filteredCards, card)
	}

	// Apply sorting
	if opts.SortBy != "" {
		// TODO: Implement sorting
	}

	// Apply pagination
	total := len(filteredCards)
	start := (opts.Page - 1) * opts.PageSize
	end := start + opts.PageSize
	if start >= total {
		start = total
	}
	if end > total {
		end = total
	}
	paginatedCards := filteredCards[start:end]

	return &types.CardSearchResult{
		Cards:    paginatedCards,
		Total:    total,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}, nil
}

// GetFilters returns the available filters for a game
func (s *SearchService) GetFilters(game string) (*types.CardFilters, error) {
	cards, err := s.cardStore.FindByGame(game)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}

	filters := &types.CardFilters{
		Sets:     make([]string, 0),
		Rarities: make([]string, 0),
	}

	// Track unique values
	setMap := make(map[string]bool)
	rarityMap := make(map[string]bool)

	for _, card := range cards {
		setMap[card.SetCode] = true
		rarityMap[card.Rarity] = true
	}

	// Convert maps to slices
	for set := range setMap {
		filters.Sets = append(filters.Sets, set)
	}
	for rarity := range rarityMap {
		filters.Rarities = append(filters.Rarities, rarity)
	}

	return filters, nil
}
