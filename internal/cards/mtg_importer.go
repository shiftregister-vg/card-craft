package cards

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/types"
)

// MTGImporter handles importing Magic: The Gathering card data
type MTGImporter struct {
	cardStore    *CardStore
	mtgCardStore *MTGCardStore
	client       *http.Client
}

// NewMTGImporter creates a new MTG importer
func NewMTGImporter(cardStore *CardStore, mtgCardStore *MTGCardStore) *MTGImporter {
	return &MTGImporter{
		cardStore:    cardStore,
		mtgCardStore: mtgCardStore,
		client: &http.Client{
			Timeout: 5 * time.Minute, // Increased timeout for bulk data download
		},
	}
}

// Import implements the Importer interface
func (i *MTGImporter) Import(ctx context.Context, store *CardStore) error {
	return i.ImportBulkData(ctx)
}

// fetchBulkDataInfo fetches information about available bulk data files
func (i *MTGImporter) fetchBulkDataInfo(ctx context.Context) (*BulkDataInfo, error) {
	url := "https://api.scryfall.com/bulk-data"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "CardCraftApp/1.0")
	req.Header.Set("Accept", "application/json;q=0.9,*/*;q=0.8")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bulk data info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch bulk data info: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data []BulkDataInfo `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find the "Default Cards" bulk data file
	for _, info := range response.Data {
		if info.Type == "default_cards" {
			return &info, nil
		}
	}

	return nil, fmt.Errorf("default_cards bulk data file not found")
}

// downloadBulkData downloads the bulk data file with retries
func (i *MTGImporter) downloadBulkData(ctx context.Context, downloadURI string) (io.ReadCloser, error) {
	maxRetries := 3
	backoff := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("Retry attempt %d after %v", retry, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		req, err := http.NewRequestWithContext(ctx, "GET", downloadURI, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "CardCraftApp/1.0")
		req.Header.Set("Accept", "application/json;q=0.9,*/*;q=0.8")

		resp, err := i.client.Do(req)
		if err != nil {
			log.Printf("Download failed (attempt %d): %v", retry+1, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("Download failed with status %d (attempt %d): %s", resp.StatusCode, retry+1, string(body))
			continue
		}

		return resp.Body, nil
	}

	return nil, fmt.Errorf("failed to download bulk data after %d attempts", maxRetries)
}

// ImportBulkData imports cards from the bulk data file
func (i *MTGImporter) ImportBulkData(ctx context.Context) error {
	startTime := time.Now()
	log.Printf("Starting import of MTG cards from bulk data")

	// Fetch bulk data info
	bulkInfo, err := i.fetchBulkDataInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch bulk data info: %w", err)
	}

	log.Printf("Found bulk data file: %s (%.2f MB)", bulkInfo.Name, float64(bulkInfo.Size)/1024/1024)

	// Download the bulk data file
	body, err := i.downloadBulkData(ctx, bulkInfo.DownloadURI)
	if err != nil {
		return fmt.Errorf("failed to download bulk data: %w", err)
	}
	defer body.Close()

	// Read the entire response into memory
	var cards []MTGAPICard
	decoder := json.NewDecoder(body)
	_, err = decoder.Token() // Read opening bracket
	if err != nil {
		return fmt.Errorf("failed to read bulk data: %w", err)
	}

	for decoder.More() {
		var card MTGAPICard
		if err := decoder.Decode(&card); err != nil {
			return fmt.Errorf("failed to decode card: %w", err)
		}
		cards = append(cards, card)
	}

	// Read closing bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read bulk data: %w", err)
	}

	log.Printf("Downloaded %d cards, starting import...", len(cards))

	// Get the last successful import timestamp
	lastImport, err := i.getLastImportTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last import timestamp: %w", err)
	}

	if lastImport != nil {
		log.Printf("Last successful import was at %s", lastImport.Format(time.RFC3339))
	}

	// Process cards in batches
	totalCards := 0
	updatedCards := 0
	newCards := 0
	skippedCards := 0
	batchSize := 500
	processedCards := 0

	for processedCards < len(cards) {
		// Process a batch of cards
		maxRetries := 5
		backoff := 5 * time.Second
		batchCards := 0
		batchTimedOut := false

		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				log.Printf("Retrying batch after %v (attempt %d/%d)", backoff, retry+1, maxRetries)
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
			}

			batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			batchCards = 0
			batchTimedOut = false

			endIdx := processedCards + batchSize
			if endIdx > len(cards) {
				endIdx = len(cards)
			}

			for _, apiCard := range cards[processedCards:endIdx] {
				select {
				case <-batchCtx.Done():
					log.Printf("Batch timeout reached after processing %d cards", batchCards)
					batchTimedOut = true
					break
				default:
					// Skip cards that haven't been updated since last import
					if lastImport != nil {
						cardUpdatedAt, err := time.Parse(time.RFC3339, apiCard.UpdatedAt)
						if err != nil {
							log.Printf("Failed to parse card update time: %v", err)
							skippedCards++
							continue
						}
						if !cardUpdatedAt.After(*lastImport) {
							skippedCards++
							continue
						}
					}

					// Check if card already exists
					existingCard, err := i.cardStore.FindByGameAndNumber("mtg", apiCard.Set, apiCard.CollectorNumber)
					if err != nil && err != sql.ErrNoRows {
						cancel()
						if retry < maxRetries-1 {
							log.Printf("Failed to check for existing card, will retry: %v", err)
							break
						}
						return fmt.Errorf("failed to check for existing card after %d retries: %w", retry+1, err)
					}

					var cardID uuid.UUID
					if existingCard != nil {
						cardID = existingCard.ID
						updatedCards++
					} else {
						cardID = uuid.New()
						newCards++
					}

					card := &types.Card{
						ID:        cardID,
						Name:      apiCard.Name,
						Game:      "mtg",
						SetCode:   apiCard.Set,
						SetName:   apiCard.SetName,
						Number:    apiCard.CollectorNumber,
						Rarity:    apiCard.Rarity,
						ImageURL:  apiCard.ImageURIs.Large,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}

					if existingCard != nil {
						card.CreatedAt = existingCard.CreatedAt
						if err := i.cardStore.Update(card); err != nil {
							log.Printf("Failed to update card %s (%s): %v", card.Name, card.Number, err)
							skippedCards++
							continue
						}
					} else {
						if err := i.cardStore.Create(card); err != nil {
							log.Printf("Failed to create card %s (%s): %v", card.Name, card.Number, err)
							skippedCards++
							continue
						}
					}

					// Create or update MTG-specific card data
					releasedAt, err := time.Parse("2006-01-02", apiCard.ReleasedAt)
					if err != nil {
						log.Printf("Failed to parse release date for card %s (%s): %v", card.Name, card.Number, err)
						skippedCards++
						continue
					}

					mtgCard := &MTGCard{
						CardID:        cardID.String(),
						ManaCost:      apiCard.ManaCost,
						CMC:           apiCard.CMC,
						TypeLine:      apiCard.TypeLine,
						OracleText:    apiCard.OracleText,
						Power:         apiCard.Power,
						Toughness:     apiCard.Toughness,
						Loyalty:       apiCard.Loyalty,
						Colors:        apiCard.Colors,
						ColorIdentity: apiCard.ColorIdentity,
						Keywords:      apiCard.Keywords,
						Legalities:    apiCard.Legalities,
						Reserved:      apiCard.Reserved,
						Foil:          apiCard.Foil,
						Nonfoil:       apiCard.Nonfoil,
						Promo:         apiCard.Promo,
						Reprint:       apiCard.Reprint,
						Variation:     apiCard.Variation,
						SetType:       apiCard.SetType,
						ReleasedAt:    releasedAt,
						CreatedAt:     time.Now(),
						UpdatedAt:     time.Now(),
					}

					existingMTGCard, err := i.mtgCardStore.FindByCardID(ctx, cardID.String())
					if err != nil {
						log.Printf("Failed to check for existing MTG card %s (%s): %v", card.Name, card.Number, err)
						skippedCards++
						continue
					}

					if existingMTGCard != nil {
						mtgCard.ID = existingMTGCard.ID
						mtgCard.CreatedAt = existingMTGCard.CreatedAt
						if err := i.mtgCardStore.Update(ctx, mtgCard); err != nil {
							log.Printf("Failed to update MTG card %s (%s): %v", card.Name, card.Number, err)
							skippedCards++
							continue
						}
					} else {
						if err := i.mtgCardStore.Create(ctx, mtgCard); err != nil {
							log.Printf("Failed to create MTG card %s (%s): %v", card.Name, card.Number, err)
							skippedCards++
							continue
						}
					}

					totalCards++
					processedCards++
					batchCards++
				}

				if batchTimedOut {
					break
				}
			}

			cancel()

			// If we processed any cards in this batch, we're done with retries
			if batchCards > 0 {
				log.Printf("Processed batch of %d cards (total: %d)", batchCards, totalCards)
				break
			}

			// If we've exhausted all retries and still haven't processed any cards, give up
			if retry == maxRetries-1 {
				return fmt.Errorf("failed to process batch after %d retries", maxRetries)
			}
		}

		// If we didn't process any cards in this batch, we're done
		if batchCards == 0 {
			break
		}
	}

	// Update the last successful import timestamp
	if err := i.updateLastImportTimestamp(ctx, time.Now()); err != nil {
		log.Printf("Failed to update last import timestamp: %v", err)
	}

	duration := time.Since(startTime)
	log.Printf("Import completed in %s", duration)
	log.Printf("Total cards processed: %d", totalCards)
	log.Printf("New cards: %d", newCards)
	log.Printf("Updated cards: %d", updatedCards)
	log.Printf("Skipped cards: %d", skippedCards)

	return nil
}

// getLastImportTimestamp retrieves the timestamp of the last successful import
func (i *MTGImporter) getLastImportTimestamp(ctx context.Context) (*time.Time, error) {
	query := `SELECT last_import FROM mtg_import_status WHERE id = 1`
	var lastImport time.Time
	err := i.cardStore.db.QueryRowContext(ctx, query).Scan(&lastImport)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last import timestamp: %w", err)
	}
	return &lastImport, nil
}

// updateLastImportTimestamp updates the timestamp of the last successful import
func (i *MTGImporter) updateLastImportTimestamp(ctx context.Context, timestamp time.Time) error {
	query := `
		INSERT INTO mtg_import_status (id, last_import)
		VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE
		SET last_import = $1
	`
	_, err := i.cardStore.db.ExecContext(ctx, query, timestamp)
	if err != nil {
		return fmt.Errorf("failed to update last import timestamp: %w", err)
	}
	return nil
}

type BulkDataInfo struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DownloadURI     string    `json:"download_uri"`
	UpdatedAt       time.Time `json:"updated_at"`
	Size            int       `json:"size"`
	ContentType     string    `json:"content_type"`
	ContentEncoding string    `json:"content_encoding"`
}

type MTGAPICard struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Set             string `json:"set"`
	SetName         string `json:"set_name"`
	CollectorNumber string `json:"collector_number"`
	Rarity          string `json:"rarity"`
	ImageURIs       struct {
		Small string `json:"small"`
		Large string `json:"large"`
	} `json:"image_uris"`
	ManaCost      string            `json:"mana_cost"`
	CMC           float64           `json:"cmc"`
	TypeLine      string            `json:"type_line"`
	OracleText    string            `json:"oracle_text"`
	Power         string            `json:"power"`
	Toughness     string            `json:"toughness"`
	Loyalty       string            `json:"loyalty"`
	Colors        []string          `json:"colors"`
	ColorIdentity []string          `json:"color_identity"`
	Keywords      []string          `json:"keywords"`
	Legalities    map[string]string `json:"legalities"`
	Reserved      bool              `json:"reserved"`
	Foil          bool              `json:"foil"`
	Nonfoil       bool              `json:"nonfoil"`
	Promo         bool              `json:"promo"`
	Reprint       bool              `json:"reprint"`
	Variation     bool              `json:"variation"`
	SetType       string            `json:"set_type"`
	ReleasedAt    string            `json:"released_at"`
	UpdatedAt     string            `json:"updated_at"`
}

// GetGame returns the game type for this importer
func (i *MTGImporter) GetGame() string {
	return "mtg"
}
