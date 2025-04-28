CREATE TABLE mtg_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    mana_cost TEXT,
    cmc DECIMAL,
    type_line TEXT,
    oracle_text TEXT,
    power TEXT,
    toughness TEXT,
    loyalty TEXT,
    colors TEXT[],
    color_identity TEXT[],
    keywords TEXT[],
    legalities JSONB,
    reserved BOOLEAN,
    foil BOOLEAN,
    nonfoil BOOLEAN,
    promo BOOLEAN,
    reprint BOOLEAN,
    variation BOOLEAN,
    set_type TEXT,
    released_at DATE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX mtg_cards_card_id_idx ON mtg_cards(card_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_mtg_cards_updated_at
    BEFORE UPDATE ON mtg_cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 