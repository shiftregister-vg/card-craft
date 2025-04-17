-- Remove indexes and constraints from deck_cards table
ALTER TABLE deck_cards DROP CONSTRAINT IF EXISTS check_quantity_positive;
ALTER TABLE deck_cards DROP CONSTRAINT IF EXISTS unique_deck_card;
ALTER TABLE deck_cards DROP CONSTRAINT IF EXISTS fk_deck_cards_card_id;
ALTER TABLE deck_cards DROP CONSTRAINT IF EXISTS fk_deck_cards_deck_id;
DROP INDEX IF EXISTS idx_deck_cards_deck_card;
DROP INDEX IF EXISTS idx_deck_cards_card_id;
DROP INDEX IF EXISTS idx_deck_cards_deck_id;

-- Remove indexes and constraints from decks table
ALTER TABLE decks DROP CONSTRAINT IF EXISTS fk_decks_user_id;
DROP INDEX IF EXISTS idx_decks_created_at;
DROP INDEX IF EXISTS idx_decks_game;
DROP INDEX IF EXISTS idx_decks_user_id;

-- Remove indexes and constraints from cards table
ALTER TABLE cards DROP CONSTRAINT IF EXISTS unique_card_game_set_number;
DROP INDEX IF EXISTS idx_cards_game_set_number;
DROP INDEX IF EXISTS idx_cards_name;
DROP INDEX IF EXISTS idx_cards_set_code;
DROP INDEX IF EXISTS idx_cards_game;

-- Remove indexes and constraints from users table
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_username;
ALTER TABLE users DROP CONSTRAINT IF EXISTS unique_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email; 