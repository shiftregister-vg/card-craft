package seed

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/database"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

type Seeder struct {
	db *sql.DB
}

func NewSeeder(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) validateTestData() error {
	// Validate users
	users := s.getTestUsers()
	for _, user := range users {
		if user.Username == "" || user.Email == "" || user.PasswordHash == "" {
			return fmt.Errorf("invalid user data: %+v", user)
		}
	}

	// Validate cards
	cards := s.getTestCards()
	for _, card := range cards {
		if card.Name == "" || card.Game == "" || card.SetCode == "" || card.Number == "" {
			return fmt.Errorf("invalid card data: %+v", card)
		}
	}

	// Validate decks
	decks := s.getTestDecks()
	for _, deck := range decks {
		if deck.Name == "" || deck.Game == "" {
			return fmt.Errorf("invalid deck data: %+v", deck)
		}
	}

	// Validate deck cards
	deckCards := s.getTestDeckCards()
	for _, dc := range deckCards {
		if dc.Quantity <= 0 {
			return fmt.Errorf("invalid deck card quantity: %d", dc.Quantity)
		}
	}

	return nil
}

func (s *Seeder) getTestUsers() []*models.User {
	return []*models.User{
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Username:     "testuser1",
			Email:        "test1@example.com",
			PasswordHash: "$2a$10$X7URVmQ7zYPHqU0vWxQ3U.3ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZ", // "password123"
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Username:     "testuser2",
			Email:        "test2@example.com",
			PasswordHash: "$2a$10$X7URVmQ7zYPHqU0vWxQ3U.3ZJZJZJZJZJZJZJZJZJZJZJZJZJZJZ", // "password123"
		},
	}
}

func (s *Seeder) getTestCards() []*models.Card {
	return []*models.Card{
		// Pokemon cards
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000101"),
			Name:     "Pikachu",
			Game:     "pokemon",
			SetCode:  "SWSH12",
			SetName:  "Silver Tempest",
			Number:   "001/195",
			Rarity:   "Rare",
			ImageUrl: "https://example.com/pikachu.jpg",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000102"),
			Name:     "Charizard",
			Game:     "pokemon",
			SetCode:  "SWSH12",
			SetName:  "Silver Tempest",
			Number:   "002/195",
			Rarity:   "Rare Holo",
			ImageUrl: "https://example.com/charizard.jpg",
		},
		// Star Wars: Unlimited cards
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			Name:     "Luke Skywalker",
			Game:     "starwars",
			SetCode:  "SWU01",
			SetName:  "Spark of Rebellion",
			Number:   "001/204",
			Rarity:   "Legendary",
			ImageUrl: "https://example.com/luke.jpg",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			Name:     "Darth Vader",
			Game:     "starwars",
			SetCode:  "SWU01",
			SetName:  "Spark of Rebellion",
			Number:   "002/204",
			Rarity:   "Legendary",
			ImageUrl: "https://example.com/vader.jpg",
		},
		// Disney Lorcana cards
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000301"),
			Name:     "Mickey Mouse",
			Game:     "lorcana",
			SetCode:  "LOR01",
			SetName:  "The First Chapter",
			Number:   "001/204",
			Rarity:   "Legendary",
			ImageUrl: "https://example.com/mickey.jpg",
		},
		{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000302"),
			Name:     "Elsa",
			Game:     "lorcana",
			SetCode:  "LOR01",
			SetName:  "The First Chapter",
			Number:   "002/204",
			Rarity:   "Legendary",
			ImageUrl: "https://example.com/elsa.jpg",
		},
	}
}

func (s *Seeder) getTestDecks() []*models.Deck {
	return []*models.Deck{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:        "My Pokemon Deck",
			Description: "A test deck for Pokemon cards",
			Game:        "pokemon",
			IsPublic:    true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:        "Star Wars Deck",
			Description: "A test deck for Star Wars cards",
			Game:        "starwars",
			IsPublic:    true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000203"),
			UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:        "Lorcana Deck",
			Description: "A test deck for Disney Lorcana cards",
			Game:        "lorcana",
			IsPublic:    false,
		},
	}
}

func (s *Seeder) getTestDeckCards() []*models.DeckCard {
	return []*models.DeckCard{
		// Pokemon deck cards
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000101"),
			Quantity: 4,
		},
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000102"),
			Quantity: 2,
		},
		// Star Wars deck cards
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			Quantity: 2,
		},
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			Quantity: 2,
		},
		// Lorcana deck cards
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000203"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000301"),
			Quantity: 4,
		},
		{
			DeckID:   uuid.MustParse("00000000-0000-0000-0000-000000000203"),
			CardID:   uuid.MustParse("00000000-0000-0000-0000-000000000302"),
			Quantity: 2,
		},
	}
}

func (s *Seeder) Seed(ctx context.Context) error {
	// Validate test data before seeding
	if err := s.validateTestData(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return database.WithTransaction(ctx, s.db, func(tx *database.Transaction) error {
		// Create test users
		for _, user := range s.getTestUsers() {
			_, err := tx.Exec(`
				INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, user.ID, user.Username, user.Email, user.PasswordHash, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}

		// Create test cards
		for _, card := range s.getTestCards() {
			_, err := tx.Exec(`
				INSERT INTO cards (id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`, card.ID, card.Name, card.Game, card.SetCode, card.SetName, card.Number, card.Rarity, card.ImageUrl, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}

		// Create test decks
		for _, deck := range s.getTestDecks() {
			_, err := tx.Exec(`
				INSERT INTO decks (id, user_id, name, description, game, is_public, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			`, deck.ID, deck.UserID, deck.Name, deck.Description, deck.Game, deck.IsPublic, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}

		// Create test deck cards
		for _, deckCard := range s.getTestDeckCards() {
			_, err := tx.Exec(`
				INSERT INTO deck_cards (deck_id, card_id, quantity, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
			`, deckCard.DeckID, deckCard.CardID, deckCard.Quantity, time.Now(), time.Now())
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Seeder) Clear(ctx context.Context) error {
	return database.WithTransaction(ctx, s.db, func(tx *database.Transaction) error {
		tables := []string{"deck_cards", "decks", "cards", "users"}
		for _, table := range tables {
			_, err := tx.Exec("DELETE FROM " + table)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
