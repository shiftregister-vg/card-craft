package cards

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type PokemonCard struct {
	ID          string       `json:"id"`
	CardID      string       `json:"card_id"`
	HP          int          `json:"hp"`
	EvolvesFrom string       `json:"evolvesFrom"`
	EvolvesTo   []string     `json:"evolvesTo"`
	Types       []string     `json:"types"`
	Subtypes    []string     `json:"subtypes"`
	Supertype   string       `json:"supertype"`
	Rules       []string     `json:"rules"`
	Abilities   []Ability    `json:"abilities"`
	Attacks     []Attack     `json:"attacks"`
	Weaknesses  []Weakness   `json:"weaknesses"`
	Resistances []Resistance `json:"resistances"`
	RetreatCost []string     `json:"retreatCost"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Ability struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Type string `json:"type"`
}

type Attack struct {
	Name                string   `json:"name"`
	Cost                []string `json:"cost"`
	ConvertedEnergyCost int      `json:"convertedEnergyCost"`
	Damage              string   `json:"damage"`
	Text                string   `json:"text"`
}

type Weakness struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Resistance struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type PokemonCardStore struct {
	db *sql.DB
}

func NewPokemonCardStore(db *sql.DB) *PokemonCardStore {
	return &PokemonCardStore{db: db}
}

func (s *PokemonCardStore) Create(ctx context.Context, card *PokemonCard) error {
	query := `
		INSERT INTO pokemon_cards (
			card_id, hp, evolves_from, evolves_to, types, subtypes,
			supertype, rules, abilities, attacks, weaknesses,
			resistances, retreat_cost
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	abilitiesJSON, err := json.Marshal(card.Abilities)
	if err != nil {
		return fmt.Errorf("failed to marshal abilities: %w", err)
	}

	attacksJSON, err := json.Marshal(card.Attacks)
	if err != nil {
		return fmt.Errorf("failed to marshal attacks: %w", err)
	}

	weaknessesJSON, err := json.Marshal(card.Weaknesses)
	if err != nil {
		return fmt.Errorf("failed to marshal weaknesses: %w", err)
	}

	resistancesJSON, err := json.Marshal(card.Resistances)
	if err != nil {
		return fmt.Errorf("failed to marshal resistances: %w", err)
	}

	err = s.db.QueryRowContext(ctx, query,
		card.CardID,
		card.HP,
		card.EvolvesFrom,
		pq.Array(card.EvolvesTo),
		pq.Array(card.Types),
		pq.Array(card.Subtypes),
		card.Supertype,
		pq.Array(card.Rules),
		abilitiesJSON,
		attacksJSON,
		weaknessesJSON,
		resistancesJSON,
		pq.Array(card.RetreatCost),
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create pokemon card: %w", err)
	}

	return nil
}

func (s *PokemonCardStore) FindByCardID(ctx context.Context, cardID string) (*PokemonCard, error) {
	query := `
		SELECT id, card_id, hp, evolves_from, evolves_to, types, subtypes,
			supertype, rules, abilities, attacks, weaknesses,
			resistances, retreat_cost, created_at, updated_at
		FROM pokemon_cards
		WHERE card_id = $1
	`

	var card PokemonCard
	var abilitiesJSON, attacksJSON, weaknessesJSON, resistancesJSON []byte

	err := s.db.QueryRowContext(ctx, query, cardID).Scan(
		&card.ID,
		&card.CardID,
		&card.HP,
		&card.EvolvesFrom,
		pq.Array(&card.EvolvesTo),
		pq.Array(&card.Types),
		pq.Array(&card.Subtypes),
		&card.Supertype,
		pq.Array(&card.Rules),
		&abilitiesJSON,
		&attacksJSON,
		&weaknessesJSON,
		&resistancesJSON,
		pq.Array(&card.RetreatCost),
		&card.CreatedAt,
		&card.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find pokemon card: %w", err)
	}

	if err := json.Unmarshal(abilitiesJSON, &card.Abilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abilities: %w", err)
	}

	if err := json.Unmarshal(attacksJSON, &card.Attacks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attacks: %w", err)
	}

	if err := json.Unmarshal(weaknessesJSON, &card.Weaknesses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weaknesses: %w", err)
	}

	if err := json.Unmarshal(resistancesJSON, &card.Resistances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resistances: %w", err)
	}

	return &card, nil
}

func (s *PokemonCardStore) Update(ctx context.Context, card *PokemonCard) error {
	query := `
		UPDATE pokemon_cards
		SET hp = $1,
			evolves_from = $2,
			evolves_to = $3,
			types = $4,
			subtypes = $5,
			supertype = $6,
			rules = $7,
			abilities = $8,
			attacks = $9,
			weaknesses = $10,
			resistances = $11,
			retreat_cost = $12,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $13
		RETURNING updated_at
	`

	abilitiesJSON, err := json.Marshal(card.Abilities)
	if err != nil {
		return fmt.Errorf("failed to marshal abilities: %w", err)
	}

	attacksJSON, err := json.Marshal(card.Attacks)
	if err != nil {
		return fmt.Errorf("failed to marshal attacks: %w", err)
	}

	weaknessesJSON, err := json.Marshal(card.Weaknesses)
	if err != nil {
		return fmt.Errorf("failed to marshal weaknesses: %w", err)
	}

	resistancesJSON, err := json.Marshal(card.Resistances)
	if err != nil {
		return fmt.Errorf("failed to marshal resistances: %w", err)
	}

	err = s.db.QueryRowContext(ctx, query,
		card.HP,
		card.EvolvesFrom,
		pq.Array(card.EvolvesTo),
		pq.Array(card.Types),
		pq.Array(card.Subtypes),
		card.Supertype,
		pq.Array(card.Rules),
		abilitiesJSON,
		attacksJSON,
		weaknessesJSON,
		resistancesJSON,
		pq.Array(card.RetreatCost),
		card.ID,
	).Scan(&card.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update pokemon card: %w", err)
	}

	return nil
}
