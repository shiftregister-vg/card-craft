-- Drop trigger first
DROP TRIGGER IF EXISTS update_decks_updated_at ON decks;

-- Drop table
DROP TABLE IF EXISTS decks;
