-- Drop indexes for deck_cards table
DROP INDEX IF EXISTS idx_deck_cards_card_id;

-- Drop indexes for decks table
DROP INDEX IF EXISTS idx_decks_game;
DROP INDEX IF EXISTS idx_decks_user_id;

-- Drop indexes for cards table
DROP INDEX IF EXISTS idx_cards_unique;
DROP INDEX IF EXISTS idx_cards_name;
DROP INDEX IF EXISTS idx_cards_game_set;

-- Drop indexes for users table
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
