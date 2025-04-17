-- Create cards table
CREATE TABLE cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    game VARCHAR(50) NOT NULL,
    set_code VARCHAR(50) NOT NULL,
    set_name VARCHAR(255) NOT NULL,
    number VARCHAR(50) NOT NULL,
    rarity VARCHAR(50) NOT NULL,
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on game and set_code for faster lookups
CREATE INDEX idx_cards_game_set ON cards(game, set_code);

-- Create index on name for faster searches
CREATE INDEX idx_cards_name ON cards(name);

-- Create unique constraint on game, set_code, and number
CREATE UNIQUE INDEX idx_cards_unique ON cards(game, set_code, number);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_cards_updated_at
    BEFORE UPDATE ON cards
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 