-- Add indexes and constraints to users table
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT unique_username UNIQUE (username);

-- Add indexes and constraints to cards table
CREATE INDEX idx_cards_game ON cards(game);
CREATE INDEX idx_cards_set_code ON cards(set_code);
CREATE INDEX idx_cards_name ON cards(name);
CREATE INDEX idx_cards_game_set_number ON cards(game, set_code, number);
ALTER TABLE cards ADD CONSTRAINT unique_card_game_set_number UNIQUE (game, set_code, number);

-- Add indexes and constraints to decks table
CREATE INDEX idx_decks_user_id ON decks(user_id);
CREATE INDEX idx_decks_game ON decks(game);
CREATE INDEX idx_decks_created_at ON decks(created_at);
ALTER TABLE decks ADD CONSTRAINT fk_decks_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Add indexes and constraints to deck_cards table
CREATE INDEX idx_deck_cards_deck_id ON deck_cards(deck_id);
CREATE INDEX idx_deck_cards_card_id ON deck_cards(card_id);
CREATE INDEX idx_deck_cards_deck_card ON deck_cards(deck_id, card_id);
ALTER TABLE deck_cards ADD CONSTRAINT fk_deck_cards_deck_id FOREIGN KEY (deck_id) REFERENCES decks(id) ON DELETE CASCADE;
ALTER TABLE deck_cards ADD CONSTRAINT fk_deck_cards_card_id FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE;
ALTER TABLE deck_cards ADD CONSTRAINT unique_deck_card UNIQUE (deck_id, card_id);
ALTER TABLE deck_cards ADD CONSTRAINT check_quantity_positive CHECK (quantity > 0); 