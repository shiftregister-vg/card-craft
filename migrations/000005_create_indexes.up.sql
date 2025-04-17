-- Add indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Add indexes for cards table
CREATE INDEX IF NOT EXISTS idx_cards_game_set ON cards(game, set_code);
CREATE INDEX IF NOT EXISTS idx_cards_name ON cards(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cards_unique ON cards(game, set_code, number);

-- Add indexes for decks table
CREATE INDEX IF NOT EXISTS idx_decks_user_id ON decks(user_id);
CREATE INDEX IF NOT EXISTS idx_decks_game ON decks(game);

-- Add indexes for deck_cards table
CREATE INDEX IF NOT EXISTS idx_deck_cards_card_id ON deck_cards(card_id);
