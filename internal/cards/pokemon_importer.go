package cards

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// PokemonImporter handles importing Pokémon card data
type PokemonImporter struct {
	cardStore    *CardStore
	pokemonStore *PokemonCardStore
	apiKey       string
}

// NewPokemonImporter creates a new Pokémon importer
func NewPokemonImporter(cardStore *CardStore, pokemonStore *PokemonCardStore) *PokemonImporter {
	apiKey := os.Getenv("POKEMON_TCG_API_KEY")
	if apiKey == "" {
		panic("POKEMON_TCG_API_KEY environment variable is not set")
	}

	return &PokemonImporter{
		cardStore:    cardStore,
		pokemonStore: pokemonStore,
		apiKey:       apiKey,
	}
}

// Import implements the Importer interface
func (i *PokemonImporter) Import(ctx context.Context, store *CardStore) error {
	return i.ImportLatestSets(ctx)
}

// fetchSets fetches available sets from the Pokemon TCG API
func (i *PokemonImporter) fetchSets(ctx context.Context) ([]PokemonSet, error) {
	url := "https://api.pokemontcg.io/v2/sets?orderBy=-releaseDate"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", i.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch sets: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []PokemonSet `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data, nil
}

// ImportSet imports all cards from a specific set
func (i *PokemonImporter) ImportSet(ctx context.Context, setID string) error {
	baseURL := "https://api.pokemontcg.io/v2/cards"
	page := 1
	totalCards := 0

	for {
		url := fmt.Sprintf("%s?q=set.id:%s&page=%d&pageSize=250", baseURL, setID, page)
		log.Printf("Fetching page %d for set %s", page, setID)

		var response PokemonResponse
		var err error

		// Retry logic with exponential backoff
		maxRetries := 3
		backoff := 2 * time.Second
		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				log.Printf("Retry attempt %d for page %d after %v", retry, page, backoff)
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
			}

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("X-Api-Key", i.apiKey)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Request failed (attempt %d): %v", retry+1, err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("Request failed with status %d (attempt %d): %s", resp.StatusCode, retry+1, string(body))
				continue
			}

			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				resp.Body.Close()
				log.Printf("Failed to decode response (attempt %d): %v", retry+1, err)
				continue
			}
			resp.Body.Close()

			// If we get here, the request was successful
			break
		}

		if err != nil {
			return fmt.Errorf("failed to fetch cards after %d retries: %w", maxRetries, err)
		}

		if len(response.Data) == 0 {
			break
		}

		log.Printf("Found %d cards on page %d", len(response.Data), page)

		for _, apiCard := range response.Data {
			// Check if card already exists
			existingCard, err := i.cardStore.FindByGameAndNumber("pokemon", apiCard.Set.ID, apiCard.Number)
			if err != nil {
				return fmt.Errorf("failed to check for existing card: %w", err)
			}

			var cardID uuid.UUID
			if existingCard != nil {
				cardID = existingCard.ID
			} else {
				cardID = uuid.New()
			}

			card := &types.Card{
				ID:        cardID,
				Name:      apiCard.Name,
				Game:      "pokemon",
				SetCode:   apiCard.Set.ID,
				SetName:   apiCard.Set.Name,
				Number:    apiCard.Number,
				Rarity:    apiCard.Rarity,
				ImageURL:  apiCard.Images.Large,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if existingCard != nil {
				card.CreatedAt = existingCard.CreatedAt
				log.Printf("Updating existing card: %s (%s) with ID: %s", card.Name, card.Number, card.ID)
				if err := i.cardStore.Update(card); err != nil {
					return fmt.Errorf("failed to update card: %w", err)
				}
			} else {
				log.Printf("Creating new card: %s (%s) with ID: %s", card.Name, card.Number, card.ID)
				if err := i.cardStore.Create(card); err != nil {
					return fmt.Errorf("failed to create card: %w", err)
				}
			}

			// Convert HP to int if it's a number
			var hp int
			if apiCard.HP != "" {
				_, err := fmt.Sscanf(apiCard.HP, "%d", &hp)
				if err != nil {
					hp = 0
				}
			}

			pokemonCard := &PokemonCard{
				CardID:      cardID.String(),
				HP:          hp,
				EvolvesFrom: apiCard.EvolvesFrom,
				EvolvesTo:   apiCard.EvolvesTo,
				Types:       apiCard.Types,
				Subtypes:    apiCard.Subtypes,
				Supertype:   apiCard.Supertype,
				Rules:       apiCard.Rules,
				Abilities:   apiCard.Abilities,
				Attacks:     apiCard.Attacks,
				Weaknesses:  apiCard.Weaknesses,
				Resistances: apiCard.Resistances,
				RetreatCost: apiCard.RetreatCost,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			existingPokemonCard, err := i.pokemonStore.FindByCardID(ctx, cardID.String())
			if err != nil {
				log.Printf("Error checking for existing pokemon card: %v", err)
				return fmt.Errorf("failed to check for existing pokemon card: %w", err)
			}

			if existingPokemonCard != nil {
				pokemonCard.ID = existingPokemonCard.ID
				pokemonCard.CreatedAt = existingPokemonCard.CreatedAt
				log.Printf("Updating existing pokemon card for: %s (%s) with CardID: %s", card.Name, card.Number, cardID)
				if err := i.pokemonStore.Update(ctx, pokemonCard); err != nil {
					log.Printf("Error updating pokemon card: %v", err)
					return fmt.Errorf("failed to update pokemon card: %w", err)
				}
			} else {
				log.Printf("Creating new pokemon card for: %s (%s) with CardID: %s", card.Name, card.Number, cardID)
				if err := i.pokemonStore.Create(ctx, pokemonCard); err != nil {
					log.Printf("Error creating pokemon card: %v", err)
					return fmt.Errorf("failed to create pokemon card: %w", err)
				}
			}

			totalCards++
		}

		page++
	}

	log.Printf("Imported %d cards from set %s", totalCards, setID)
	return nil
}

// ImportLatestSets imports cards from the latest Pokémon sets
func (i *PokemonImporter) ImportLatestSets(ctx context.Context) error {
	startTime := time.Now()
	log.Printf("Starting import of Pokemon sets")

	// Fetch available sets from the API
	sets, err := i.fetchSets(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch sets: %w", err)
	}

	log.Printf("Found %d sets to import", len(sets))

	// Import each set
	for _, set := range sets {
		log.Printf("Importing set %s (%s)", set.Name, set.ID)
		if err := i.ImportSet(ctx, set.ID); err != nil {
			return fmt.Errorf("failed to import set %s: %w", set.ID, err)
		}
	}

	duration := time.Since(startTime)
	log.Printf("Import completed in %s", duration)
	return nil
}

type PokemonResponse struct {
	Data []PokemonAPICard `json:"data"`
}

type PokemonSet struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Series       string `json:"series"`
	PrintedTotal int    `json:"printedTotal"`
	Total        int    `json:"total"`
	ReleaseDate  string `json:"releaseDate"`
}

type PokemonAPICard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Set  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"set"`
	Number string `json:"number"`
	Rarity string `json:"rarity"`
	Images struct {
		Large string `json:"large"`
		Small string `json:"small"`
	} `json:"images"`
	HP          string   `json:"hp"`
	EvolvesFrom string   `json:"evolvesFrom"`
	EvolvesTo   []string `json:"evolvesTo"`
	Types       []string `json:"types"`
	Subtypes    []string `json:"subtypes"`
	Supertype   string   `json:"supertype"`
	Rules       []string `json:"rules"`
	Abilities   []Ability
	Attacks     []Attack
	Weaknesses  []Weakness
	Resistances []Resistance
	RetreatCost []string `json:"retreatCost"`
}

// GetGame returns the game type for this importer
func (i *PokemonImporter) GetGame() string {
	return "pokemon"
}
