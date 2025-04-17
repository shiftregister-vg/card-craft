-- Drop trigger first
DROP TRIGGER IF EXISTS update_deck_cards_updated_at ON deck_cards;

-- Drop table
DROP TABLE IF EXISTS deck_cards;
