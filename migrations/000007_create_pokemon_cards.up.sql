CREATE TABLE IF NOT EXISTS pokemon_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    hp INTEGER,
    evolves_from TEXT,
    evolves_to TEXT[],
    types TEXT[],
    subtypes TEXT[],
    supertype TEXT,
    rules TEXT[],
    abilities JSONB,
    attacks JSONB,
    weaknesses JSONB,
    resistances JSONB,
    retreat_cost TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_pokemon_cards_card_id FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pokemon_cards_card_id ON pokemon_cards(card_id); 