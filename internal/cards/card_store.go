package cards

import (
	"context"

	"github.com/shiftregister-vg/card-craft/internal/models"
)

// GetCard retrieves a card by ID
func (s *CardStore) GetCard(ctx context.Context, id string) (*models.Card, error) {
	var card models.Card
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at
		FROM cards
		WHERE id = $1
	`, id).Scan(
		&card.ID,
		&card.Name,
		&card.Game,
		&card.SetCode,
		&card.SetName,
		&card.Number,
		&card.Rarity,
		&card.ImageUrl,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}
