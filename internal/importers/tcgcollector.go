package importers

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

type TCGCollectorImporter struct {
	cardStore *models.CardStore
}

func NewTCGCollectorImporter(cardStore *models.CardStore) *TCGCollectorImporter {
	return &TCGCollectorImporter{
		cardStore: cardStore,
	}
}

func (i *TCGCollectorImporter) Import(reader io.Reader) (*models.ImportResult, error) {
	csvReader := csv.NewReader(reader)

	// Skip header row
	if _, err := csvReader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	result := &models.ImportResult{
		TotalCards:    0,
		ImportedCards: 0,
		UpdatedCards:  0,
		Errors:        []string{},
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error reading row: %v", err))
			continue
		}

		result.TotalCards++

		// TCGCollector CSV format:
		// Name, Set Name, Set Code, Number, Rarity, Game
		if len(record) < 6 {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid row format: %v", record))
			continue
		}

		card := &models.Card{
			ID:      uuid.New(),
			Name:    strings.TrimSpace(record[0]),
			SetName: strings.TrimSpace(record[1]),
			SetCode: strings.TrimSpace(record[2]),
			Number:  strings.TrimSpace(record[3]),
			Rarity:  strings.TrimSpace(record[4]),
			Game:    strings.TrimSpace(record[5]),
		}

		// Check if card already exists
		existingCard, err := i.cardStore.FindByGameSetAndNumber(card.Game, card.SetCode, card.Number)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error checking for existing card: %v", err))
			continue
		}

		if existingCard != nil {
			// Update existing card
			existingCard.Name = card.Name
			existingCard.SetName = card.SetName
			existingCard.Rarity = card.Rarity

			if err := i.cardStore.Update(existingCard); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error updating card: %v", err))
				continue
			}
			result.UpdatedCards++
		} else {
			// Create new card
			if err := i.cardStore.Create(card); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error creating card: %v", err))
				continue
			}
			result.ImportedCards++
		}
	}

	return result, nil
}
