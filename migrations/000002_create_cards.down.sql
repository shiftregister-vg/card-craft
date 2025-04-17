-- Drop trigger first
DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;

-- Drop table
DROP TABLE IF EXISTS cards;
