package search

import (
	"strings"

	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// SearchService handles card search operations
type SearchService struct {
	cardStore *cards.CardStore
}

// NewSearchService creates a new search service
func NewSearchService(cardStore *cards.CardStore) *SearchService {
	return &SearchService{
		cardStore: cardStore,
	}
}

// Search searches for cards based on the provided options
func (s *SearchService) Search(opts types.SearchOptions) (*types.CardSearchResult, error) {
	// If we're searching by set code and number, use FindByGameAndNumber
	if opts.SetCode != "" && opts.Name != "" {
		card, err := s.cardStore.FindByGameAndNumber(opts.Game, opts.SetCode, opts.Name)
		if err != nil {
			return nil, err
		}
		if card != nil {
			return &types.CardSearchResult{
				Cards:      []*types.Card{card},
				TotalCount: 1,
				Page:       1,
				PageSize:   1,
			}, nil
		}
		return &types.CardSearchResult{
			Cards:      []*types.Card{},
			TotalCount: 0,
			Page:       1,
			PageSize:   1,
		}, nil
	}

	// Use the database's search capabilities
	game := opts.Game
	name := opts.Name
	if name == "" {
		// If no name is provided, get all cards for the game
		cards, _, err := s.cardStore.FindByGame(game, 50, "")
		if err != nil {
			return nil, err
		}
		return &types.CardSearchResult{
			Cards:      cards,
			TotalCount: len(cards),
			Page:       1,
			PageSize:   50,
		}, nil
	}

	// Search by name
	cards, err := s.cardStore.SearchCards(name, game)
	if err != nil {
		return nil, err
	}

	// Apply additional filters if needed
	var filteredCards []*types.Card
	for _, card := range cards {
		// Filter by set code
		if opts.SetCode != "" && !strings.EqualFold(card.SetCode, opts.SetCode) {
			continue
		}

		// Filter by rarity
		if opts.Rarity != "" && !strings.EqualFold(card.Rarity, opts.Rarity) {
			continue
		}

		filteredCards = append(filteredCards, card)
	}

	// Calculate pagination
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 50
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(filteredCards) {
		end = len(filteredCards)
	}

	var pagedCards []*types.Card
	if start < len(filteredCards) {
		pagedCards = filteredCards[start:end]
	}

	return &types.CardSearchResult{
		Cards:      pagedCards,
		TotalCount: len(filteredCards),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// FindByGameAndNumber finds a card by its game, set code, and number
func (s *SearchService) FindByGameAndNumber(game, setCode, number string) (*types.Card, error) {
	return s.cardStore.FindByGameAndNumber(game, setCode, number)
}
