-- Drop the old primary key constraint
ALTER TABLE deck_cards DROP CONSTRAINT deck_cards_pkey;

-- Add id column to deck_cards table
ALTER TABLE deck_cards ADD COLUMN id UUID DEFAULT gen_random_uuid();

-- Add a unique constraint on deck_id and card_id
ALTER TABLE deck_cards ADD CONSTRAINT deck_cards_deck_id_card_id_unique UNIQUE (deck_id, card_id);

-- Make id the primary key
ALTER TABLE deck_cards ADD PRIMARY KEY (id); 