package cards

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	// Check for cached version first
	cacheDir := os.Getenv("DEVBOX_PROJECT_ROOT") + "/.devbox/cache"
	cacheFile := cacheDir + "/mtg_bulk_data.json"

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("Failed to create cache directory: %v", err)
	} else {
		// Check if we have a valid cached version
		if info, err := os.Stat(cacheFile); err == nil {
			// Check if cache is less than 24 hours old
			if time.Since(info.ModTime()) < 24*time.Hour {
				// Read the cached file
				file, err := os.Open(cacheFile)
				if err == nil {
					log.Printf("Using cached bulk data file from %s", info.ModTime().Format(time.RFC3339))
					return file, nil
				}
				log.Printf("Failed to open cached file: %v", err)
			} else {
				log.Printf("Cache is older than 24 hours, downloading fresh data")
			}
		}
	}

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

		// Create a temporary file to store the response
		tempFile, err := os.CreateTemp(cacheDir, "mtg_bulk_data_*.json")
		if err != nil {
			resp.Body.Close()
			log.Printf("Failed to create temporary file: %v", err)
			continue
		}

		// Copy the response to the temporary file
		_, err = io.Copy(tempFile, resp.Body)
		resp.Body.Close()
		if err != nil {
			tempFile.Close()
			os.Remove(tempFile.Name())
			log.Printf("Failed to write to temporary file: %v", err)
			continue
		}
		tempFile.Close()

		// Rename the temporary file to the cache file
		if err := os.Rename(tempFile.Name(), cacheFile); err != nil {
			os.Remove(tempFile.Name())
			log.Printf("Failed to rename temporary file: %v", err)
			continue
		}

		// Open the cache file for reading
		file, err := os.Open(cacheFile)
		if err != nil {
			log.Printf("Failed to open cache file: %v", err)
			continue
		}

		return file, nil
	}

	return nil, fmt.Errorf("failed to download bulk data after %d attempts", maxRetries)
}

// ImportBulkData imports cards from the bulk data file
func (i *MTGImporter) ImportBulkData(ctx context.Context) error {
	// Get last import timestamp
	lastImport, err := i.getLastImportTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last import timestamp: %w", err)
	}

	log.Printf("Starting import. Last import was at: %v", lastImport)

	// Get bulk data info
	bulkData, err := i.fetchBulkDataInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bulk data info: %w", err)
	}

	log.Printf("Found bulk data file: %s (%.2f MB)", bulkData.Name, float64(bulkData.Size)/1024/1024)

	// Download bulk data
	body, err := i.downloadBulkData(ctx, bulkData.DownloadURI)
	if err != nil {
		return fmt.Errorf("failed to download bulk data: %w", err)
	}
	defer body.Close()

	// Process cards in batches
	batchSize := 100
	var batch []MTGAPICard
	processed := 0

	// Read and process cards in batches
	decoder := json.NewDecoder(body)

	// Read the opening bracket
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read opening bracket: %w", err)
	}

	for decoder.More() {
		var card MTGAPICard
		if err := decoder.Decode(&card); err != nil {
			return fmt.Errorf("failed to decode card: %w", err)
		}

		if lastImport != nil {
			cardUpdatedAt, err := time.Parse(time.RFC3339, card.UpdatedAt)
			if err == nil && !cardUpdatedAt.After(*lastImport) {
				continue
			}
		}

		batch = append(batch, card)

		// Process batch when it reaches batchSize
		if len(batch) >= batchSize {
			if err := i.processBatch(ctx, batch, lastImport); err != nil {
				return fmt.Errorf("failed to process batch: %w", err)
			}
			processed += len(batch)
			log.Printf("Processed %d cards", processed)
			batch = nil
		}
	}

	// Process remaining cards
	if len(batch) > 0 {
		if err := i.processBatch(ctx, batch, lastImport); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
		processed += len(batch)
		log.Printf("Processed %d cards", processed)
	}

	// Read the closing bracket
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read closing bracket: %w", err)
	}

	// Update last import timestamp
	if err := i.updateLastImportTimestamp(ctx, time.Now()); err != nil {
		return fmt.Errorf("failed to update last import timestamp: %w", err)
	}

	log.Printf("Import completed successfully. Total cards processed: %d", processed)
	return nil
}

func (i *MTGImporter) processBatch(ctx context.Context, batch []MTGAPICard, lastImport *time.Time) error {
	for _, card := range batch {
		// Skip cards that haven't been updated since last import
		var updatedAt time.Time
		var err error
		if card.UpdatedAt != "" {
			updatedAt, err = time.Parse(time.RFC3339, card.UpdatedAt)
			if err != nil {
				return fmt.Errorf("failed to parse updated_at timestamp: %w", err)
			}
		} else {
			// If no updated_at timestamp is provided, use current time
			updatedAt = time.Now()
		}

		if lastImport != nil && !updatedAt.After(*lastImport) {
			continue
		}

		// Check if card exists
		existingCard, err := i.cardStore.FindByGameAndNumber("mtg", card.Set, card.CollectorNumber)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check for existing card: %w", err)
		}

		cardPtr := &card
		if err == sql.ErrNoRows {
			if err := i.createCard(ctx, cardPtr); err != nil {
				return fmt.Errorf("failed to create card: %w", err)
			}
		} else {
			if err := i.updateCard(ctx, existingCard.ID, cardPtr); err != nil {
				return fmt.Errorf("failed to update card: %w", err)
			}
		}
	}
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

func (i *MTGImporter) updateCard(ctx context.Context, cardID uuid.UUID, apiCard *MTGAPICard) error {
	// Update base card
	card := &types.Card{
		ID:        cardID,
		Name:      apiCard.Name,
		Game:      "mtg",
		SetCode:   apiCard.Set,
		SetName:   apiCard.SetName,
		Number:    apiCard.CollectorNumber,
		Rarity:    apiCard.Rarity,
		ImageURL:  apiCard.ImageURIs.Large,
		UpdatedAt: time.Now(),
	}

	if err := i.cardStore.Update(card); err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	// Update MTG-specific card data
	releasedAt, err := time.Parse("2006-01-02", apiCard.ReleasedAt)
	if err != nil {
		return fmt.Errorf("failed to parse release date: %w", err)
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
		UpdatedAt:     time.Now(),
	}

	return i.mtgCardStore.Update(ctx, mtgCard)
}

func (i *MTGImporter) createCard(ctx context.Context, apiCard *MTGAPICard) error {
	// Create base card
	card := &types.Card{
		ID:        uuid.New(),
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

	if err := i.cardStore.Create(card); err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	// Create MTG-specific card data
	releasedAt, err := time.Parse("2006-01-02", apiCard.ReleasedAt)
	if err != nil {
		return fmt.Errorf("failed to parse release date: %w", err)
	}

	mtgCard := &MTGCard{
		CardID:        card.ID.String(),
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

	return i.mtgCardStore.Create(ctx, mtgCard)
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
