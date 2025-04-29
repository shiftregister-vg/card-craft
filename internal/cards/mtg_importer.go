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
	"strings"
	"sync"
	"time"

	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shiftregister-vg/card-craft/internal/database"
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

// ImportBulkData imports all cards from the bulk data
func (i *MTGImporter) ImportBulkData(ctx context.Context) error {
	startTime := time.Now()
	log.Printf("Starting MTG card import process")

	// Get bulk data info
	info, err := i.fetchBulkDataInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch bulk data info: %w", err)
	}

	// Get the last import timestamp
	lastImport, err := i.getLastImportTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last import timestamp: %w", err)
	}

	// Download and process the bulk data
	downloadStart := time.Now()
	reader, err := i.downloadBulkData(ctx, info.DownloadURI)
	if err != nil {
		return fmt.Errorf("failed to download bulk data: %w", err)
	}
	defer reader.Close()
	log.Printf("Downloaded bulk data in %v", time.Since(downloadStart))

	// Create a decoder for the JSON data
	decoder := json.NewDecoder(reader)

	// Skip the opening array token
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read opening array token: %w", err)
	}

	// Constants for batch processing
	const (
		batchSize     = 100
		numWorkers    = 8 // Optimal number of workers based on typical CPU cores
		errorChanSize = 100
	)

	// Build all batches first
	buildStart := time.Now()
	var batches [][]MTGAPICard
	var currentBatch []MTGAPICard
	var processedCards int

	log.Printf("Building batches...")
	for decoder.More() {
		var card MTGAPICard
		if err := decoder.Decode(&card); err != nil {
			return fmt.Errorf("failed to decode card: %w", err)
		}

		currentBatch = append(currentBatch, card)
		processedCards++

		if len(currentBatch) >= batchSize {
			batches = append(batches, currentBatch)
			currentBatch = nil
		}

		// Log progress every 1000 cards
		if processedCards%1000 == 0 {
			log.Printf("Built %d cards into %d batches", processedCards, len(batches))
		}
	}

	// Add the final batch if it has any cards
	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	buildDuration := time.Since(buildStart)
	log.Printf("Finished building %d batches with %d total cards in %v",
		len(batches), processedCards, buildDuration)

	// Create channels for batch processing
	batchChan := make(chan []MTGAPICard, numWorkers*2)
	errorChan := make(chan error, errorChanSize)
	doneChan := make(chan struct{})
	progressChan := make(chan int, numWorkers*2)                          // Channel for tracking progress
	statsChan := make(chan struct{ inserted, updated int }, numWorkers*2) // Channel for tracking stats

	// Create a wait group for workers
	var wg sync.WaitGroup

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				inserted, updated, err := i.processBatch(ctx, batch, lastImport)
				if err != nil {
					select {
					case errorChan <- fmt.Errorf("failed to process batch: %w", err):
					default:
						// If error channel is full, log and continue
						log.Printf("Error processing batch: %v", err)
					}
				}
				// Report batch completion and stats
				select {
				case progressChan <- 1:
				default:
					// If progress channel is full, skip reporting
				}
				select {
				case statsChan <- struct{ inserted, updated int }{inserted, updated}:
				default:
					// If stats channel is full, skip reporting
				}
			}
		}()
	}

	// Start a goroutine to close doneChan when all workers are done
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Start a goroutine to track progress and stats
	var processedBatches int
	var totalInserted, totalUpdated int
	go func() {
		for {
			select {
			case <-progressChan:
				processedBatches++
				if processedBatches%10 == 0 { // Log every 10 batches
					log.Printf("Processed %d/%d batches (%.1f%%)",
						processedBatches, len(batches),
						float64(processedBatches)/float64(len(batches))*100)
				}
			case stats := <-statsChan:
				totalInserted += stats.inserted
				totalUpdated += stats.updated
			case <-doneChan:
				return
			}
		}
	}()

	// Start processing batches
	processStart := time.Now()
	log.Printf("Starting batch processing with %d workers...", numWorkers)

	// Send all batches to the workers
	for _, batch := range batches {
		select {
		case batchChan <- batch:
		case err := <-errorChan:
			return fmt.Errorf("worker error: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Close the batch channel to signal workers to finish
	close(batchChan)

	// Wait for all workers to complete
	select {
	case <-doneChan:
		// All workers completed successfully
	case err := <-errorChan:
		return fmt.Errorf("worker error: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}

	processDuration := time.Since(processStart)
	log.Printf("Batch processing completed in %v", processDuration)

	// Update the last import timestamp
	if err := i.updateLastImportTimestamp(ctx, time.Now()); err != nil {
		return fmt.Errorf("failed to update last import timestamp: %w", err)
	}

	totalDuration := time.Since(startTime)
	log.Printf("Import completed in %v", totalDuration)
	log.Printf("Summary:")
	log.Printf("  - Download time: %v", downloadStart.Sub(startTime))
	log.Printf("  - Batch building time: %v", buildDuration)
	log.Printf("  - Batch processing time: %v", processDuration)
	log.Printf("  - Total time: %v", totalDuration)
	log.Printf("  - Total cards processed: %d", processedCards)
	log.Printf("  - Total cards inserted: %d", totalInserted)
	log.Printf("  - Total cards updated: %d", totalUpdated)
	log.Printf("  - Total batches: %d", len(batches))
	log.Printf("  - Average time per batch: %v", processDuration/time.Duration(len(batches)))
	log.Printf("  - Average time per card: %v", totalDuration/time.Duration(processedCards))

	return nil
}

// CardSignature represents the fields that determine if a card needs updating
type CardSignature struct {
	Name       string            `json:"name"`
	SetName    string            `json:"setName"`
	Rarity     string            `json:"rarity"`
	ImageURL   string            `json:"imageUrl"`
	ManaCost   string            `json:"manaCost"`
	TypeLine   string            `json:"typeLine"`
	OracleText string            `json:"oracleText"`
	Power      string            `json:"power"`
	Toughness  string            `json:"toughness"`
	Loyalty    string            `json:"loyalty"`
	Colors     []string          `json:"colors"`
	Keywords   []string          `json:"keywords"`
	Legalities map[string]string `json:"legalities"`
}

func (c *CardSignature) Hash() string {
	data, _ := json.Marshal(c)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (i *MTGImporter) processBatch(ctx context.Context, batch []MTGAPICard, lastImport *time.Time) (int, int, error) {
	// First, check which cards exist and which need to be created
	var cardsToCreate []*types.Card
	var cardsToUpdate []*types.Card
	var mtgCardsToCreate []*MTGCard
	var mtgCardsToUpdate []*MTGCard

	// Build a map of set+number to card for quick lookup
	cardMap := make(map[string]*types.Card)
	missingMtgCards := make(map[string]bool) // Track cards with missing mtg_cards records
	for _, card := range batch {
		if card.Set == "" || card.CollectorNumber == "" {
			log.Printf("Warning: skipping card with missing set or collector number")
			continue
		}
		key := fmt.Sprintf("%s:%s", card.Set, card.CollectorNumber)
		cardMap[key] = nil // Initialize with nil, will be populated if card exists
	}

	// Fetch all existing cards in a single query
	if len(cardMap) > 0 {
		query := `
			SELECT c.id, c.name, c.game, c.set_code, c.set_name, c.number, c.rarity, c.image_url, c.created_at, c.updated_at,
				       m.id as mtg_id, m.mana_cost, m.cmc, m.type_line, m.oracle_text, m.power, m.toughness, m.loyalty,
				       m.colors, m.color_identity, m.keywords, m.legalities, m.reserved, m.foil, m.nonfoil,
				       m.promo, m.reprint, m.variation, m.set_type, m.released_at
			FROM cards c
			LEFT JOIN mtg_cards m ON c.id = m.card_id
			WHERE c.game = 'mtg' AND (c.set_code, c.number) IN (
				SELECT unnest($1::text[]), unnest($2::text[])
			)
		`
		var setCodes, numbers []string
		for key := range cardMap {
			parts := strings.Split(key, ":")
			setCodes = append(setCodes, parts[0])
			numbers = append(numbers, parts[1])
		}

		rows, err := i.cardStore.db.QueryContext(ctx, query, pq.Array(setCodes), pq.Array(numbers))
		if err != nil {
			return 0, 0, fmt.Errorf("failed to query existing cards: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var card types.Card
			var mtgCard MTGCard
			var legalitiesJSON []byte
			var manaCost, typeLine, oracleText, power, toughness, loyalty sql.NullString
			var cmc sql.NullFloat64
			var colors, colorIdentity, keywords []string
			var reserved, foil, nonfoil, promo, reprint, variation sql.NullBool
			var setType sql.NullString
			var releasedAt sql.NullTime
			var mtgID sql.NullString

			err := rows.Scan(
				&card.ID,
				&card.Name,
				&card.Game,
				&card.SetCode,
				&card.SetName,
				&card.Number,
				&card.Rarity,
				&card.ImageURL,
				&card.CreatedAt,
				&card.UpdatedAt,
				&mtgID,
				&manaCost,
				&cmc,
				&typeLine,
				&oracleText,
				&power,
				&toughness,
				&loyalty,
				pq.Array(&colors),
				pq.Array(&colorIdentity),
				pq.Array(&keywords),
				&legalitiesJSON,
				&reserved,
				&foil,
				&nonfoil,
				&promo,
				&reprint,
				&variation,
				&setType,
				&releasedAt,
			)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to scan card: %w", err)
			}

			// If mtg_id is null, it means the mtg_cards record was deleted
			if !mtgID.Valid {
				key := fmt.Sprintf("%s:%s", card.SetCode, card.Number)
				cardMap[key] = &card        // Keep the existing card
				missingMtgCards[key] = true // Mark as having missing mtg_cards record
				log.Printf("Found card %s (%s) with missing mtg_cards record", card.Name, card.Number)
				continue
			}

			// Convert nullable fields to their proper types
			mtgCard.ID = mtgID.String
			mtgCard.CardID = card.ID.String() // Set the CardID to link to the existing card
			mtgCard.ManaCost = manaCost.String
			mtgCard.CMC = cmc.Float64
			mtgCard.TypeLine = typeLine.String
			mtgCard.OracleText = oracleText.String
			mtgCard.Power = power.String
			mtgCard.Toughness = toughness.String
			mtgCard.Loyalty = loyalty.String
			mtgCard.Colors = colors
			mtgCard.ColorIdentity = colorIdentity
			mtgCard.Keywords = keywords
			mtgCard.Reserved = reserved.Bool
			mtgCard.Foil = foil.Bool
			mtgCard.Nonfoil = nonfoil.Bool
			mtgCard.Promo = promo.Bool
			mtgCard.Reprint = reprint.Bool
			mtgCard.Variation = variation.Bool
			mtgCard.SetType = setType.String
			mtgCard.ReleasedAt = releasedAt.Time

			if legalitiesJSON != nil {
				if err := json.Unmarshal(legalitiesJSON, &mtgCard.Legalities); err != nil {
					return 0, 0, fmt.Errorf("failed to unmarshal legalities: %w", err)
				}
			}

			key := fmt.Sprintf("%s:%s", card.SetCode, card.Number)
			cardMap[key] = &card
		}
	}

	// Process each card in the batch
	for _, card := range batch {
		// Skip invalid cards
		if card.Set == "" || card.CollectorNumber == "" {
			continue
		}

		// Skip cards that haven't been updated since last import
		var updatedAt time.Time
		var err error
		if card.UpdatedAt != "" {
			updatedAt, err = time.Parse(time.RFC3339, card.UpdatedAt)
			if err != nil {
				log.Printf("Warning: failed to parse updated_at timestamp for card %s: %v", card.Name, err)
				updatedAt = time.Now()
			}
		} else {
			updatedAt = time.Now()
		}

		if lastImport != nil && !updatedAt.After(*lastImport) {
			continue
		}

		key := fmt.Sprintf("%s:%s", card.Set, card.CollectorNumber)
		existingCard := cardMap[key]

		// Create base card
		baseCard := &types.Card{
			Name:      card.Name,
			Game:      "mtg",
			SetCode:   card.Set,
			SetName:   card.SetName,
			Number:    card.CollectorNumber,
			Rarity:    card.Rarity,
			ImageURL:  card.ImageURIs.Large,
			UpdatedAt: time.Now(),
		}

		// Create MTG-specific card
		releasedAt, err := time.Parse("2006-01-02", card.ReleasedAt)
		if err != nil {
			log.Printf("Warning: failed to parse release date for card %s: %v", card.Name, err)
			releasedAt = time.Now()
		}

		mtgCard := &MTGCard{
			ManaCost:      card.ManaCost,
			CMC:           card.CMC,
			TypeLine:      card.TypeLine,
			OracleText:    card.OracleText,
			Power:         card.Power,
			Toughness:     card.Toughness,
			Loyalty:       card.Loyalty,
			Colors:        card.Colors,
			ColorIdentity: card.ColorIdentity,
			Keywords:      card.Keywords,
			Legalities:    card.Legalities,
			Reserved:      card.Reserved,
			Foil:          card.Foil,
			Nonfoil:       card.Nonfoil,
			Promo:         card.Promo,
			Reprint:       card.Reprint,
			Variation:     card.Variation,
			SetType:       card.SetType,
			ReleasedAt:    releasedAt,
			UpdatedAt:     time.Now(),
		}

		if existingCard == nil {
			// Card doesn't exist, prepare for creation
			baseCard.ID = uuid.New()
			baseCard.CreatedAt = time.Now()
			mtgCard.CardID = baseCard.ID.String()
			mtgCard.CreatedAt = time.Now()

			cardsToCreate = append(cardsToCreate, baseCard)
			mtgCardsToCreate = append(mtgCardsToCreate, mtgCard)
		} else {
			// Check if we need to create the mtg_cards record
			if missingMtgCards[key] {
				// MTG card record is missing, prepare for creation
				mtgCard.CardID = existingCard.ID.String()
				mtgCard.CreatedAt = time.Now()
				log.Printf("Preparing to create missing mtg_cards record for %s (%s)", existingCard.Name, existingCard.Number)
				mtgCardsToCreate = append(mtgCardsToCreate, mtgCard)
				continue
			}

			// Create signature for the new card
			newSignature := &CardSignature{
				Name:       baseCard.Name,
				SetName:    baseCard.SetName,
				Rarity:     baseCard.Rarity,
				ImageURL:   baseCard.ImageURL,
				ManaCost:   card.ManaCost,
				TypeLine:   card.TypeLine,
				OracleText: card.OracleText,
				Power:      card.Power,
				Toughness:  card.Toughness,
				Loyalty:    card.Loyalty,
				Colors:     card.Colors,
				Keywords:   card.Keywords,
				Legalities: card.Legalities,
			}

			// Create signature for the existing card
			existingSignature := &CardSignature{
				Name:       existingCard.Name,
				SetName:    existingCard.SetName,
				Rarity:     existingCard.Rarity,
				ImageURL:   existingCard.ImageURL,
				ManaCost:   mtgCard.ManaCost,
				TypeLine:   mtgCard.TypeLine,
				OracleText: mtgCard.OracleText,
				Power:      mtgCard.Power,
				Toughness:  mtgCard.Toughness,
				Loyalty:    mtgCard.Loyalty,
				Colors:     mtgCard.Colors,
				Keywords:   mtgCard.Keywords,
				Legalities: mtgCard.Legalities,
			}

			// Compare hashes to determine if update is needed
			if newSignature.Hash() != existingSignature.Hash() {
				// Card exists and has changed, prepare for update
				baseCard.ID = existingCard.ID
				baseCard.CreatedAt = existingCard.CreatedAt
				mtgCard.CardID = existingCard.ID.String()
				mtgCard.CreatedAt = existingCard.CreatedAt

				cardsToUpdate = append(cardsToUpdate, baseCard)
				mtgCardsToUpdate = append(mtgCardsToUpdate, mtgCard)
			}
		}
	}

	// Perform batch operations in a transaction
	err := database.WithTransaction(ctx, i.cardStore.db, func(tx *database.Transaction) error {
		if len(cardsToCreate) > 0 {
			log.Printf("Creating %d new cards", len(cardsToCreate))
			// Create base cards
			query := `
				INSERT INTO cards (id, name, game, set_code, set_name, number, rarity, image_url, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			`
			for _, card := range cardsToCreate {
				_, err := tx.Exec(query,
					card.ID,
					card.Name,
					card.Game,
					card.SetCode,
					card.SetName,
					card.Number,
					card.Rarity,
					card.ImageURL,
					card.CreatedAt,
					card.UpdatedAt,
				)
				if err != nil {
					return fmt.Errorf("failed to create card: %w", err)
				}
			}
		}

		if len(mtgCardsToCreate) > 0 {
			log.Printf("Creating %d new MTG cards", len(mtgCardsToCreate))
			// Create MTG cards
			mtgCreateQuery := `
				INSERT INTO mtg_cards (
					card_id, mana_cost, cmc, type_line, oracle_text, power, toughness, loyalty,
					colors, color_identity, keywords, legalities, reserved, foil, nonfoil,
					promo, reprint, variation, set_type, released_at, created_at, updated_at
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
					$16, $17, $18, $19, $20, $21, $22
				)
			`
			for _, card := range mtgCardsToCreate {
				legalitiesJSON, err := json.Marshal(card.Legalities)
				if err != nil {
					return fmt.Errorf("failed to marshal legalities: %w", err)
				}

				_, err = tx.Exec(mtgCreateQuery,
					card.CardID,
					card.ManaCost,
					card.CMC,
					card.TypeLine,
					card.OracleText,
					card.Power,
					card.Toughness,
					card.Loyalty,
					pq.Array(card.Colors),
					pq.Array(card.ColorIdentity),
					pq.Array(card.Keywords),
					legalitiesJSON,
					card.Reserved,
					card.Foil,
					card.Nonfoil,
					card.Promo,
					card.Reprint,
					card.Variation,
					card.SetType,
					card.ReleasedAt,
					card.CreatedAt,
					card.UpdatedAt,
				)
				if err != nil {
					return fmt.Errorf("failed to create MTG card: %w", err)
				}
			}
		}

		if len(cardsToUpdate) > 0 {
			log.Printf("Updating %d existing cards", len(cardsToUpdate))
			// Update base cards
			updateQuery := `
				UPDATE cards
				SET name = $1, game = $2, set_code = $3, set_name = $4, number = $5, rarity = $6, image_url = $7, updated_at = $8
				WHERE id = $9
			`
			for _, card := range cardsToUpdate {
				_, err := tx.Exec(updateQuery,
					card.Name,
					card.Game,
					card.SetCode,
					card.SetName,
					card.Number,
					card.Rarity,
					card.ImageURL,
					card.UpdatedAt,
					card.ID,
				)
				if err != nil {
					return fmt.Errorf("failed to update card: %w", err)
				}
			}
		}

		if len(mtgCardsToUpdate) > 0 {
			log.Printf("Updating %d existing MTG cards", len(mtgCardsToUpdate))
			// Update MTG cards
			mtgUpdateQuery := `
				UPDATE mtg_cards
				SET mana_cost = $1, cmc = $2, type_line = $3, oracle_text = $4, power = $5, toughness = $6, loyalty = $7,
					colors = $8, color_identity = $9, keywords = $10, legalities = $11, reserved = $12, foil = $13, nonfoil = $14,
					promo = $15, reprint = $16, variation = $17, set_type = $18, released_at = $19, updated_at = $20
				WHERE card_id = $21
			`
			for _, card := range mtgCardsToUpdate {
				legalitiesJSON, err := json.Marshal(card.Legalities)
				if err != nil {
					return fmt.Errorf("failed to marshal legalities: %w", err)
				}

				_, err = tx.Exec(mtgUpdateQuery,
					card.ManaCost,
					card.CMC,
					card.TypeLine,
					card.OracleText,
					card.Power,
					card.Toughness,
					card.Loyalty,
					pq.Array(card.Colors),
					pq.Array(card.ColorIdentity),
					pq.Array(card.Keywords),
					legalitiesJSON,
					card.Reserved,
					card.Foil,
					card.Nonfoil,
					card.Promo,
					card.Reprint,
					card.Variation,
					card.SetType,
					card.ReleasedAt,
					card.UpdatedAt,
					card.CardID,
				)
				if err != nil {
					return fmt.Errorf("failed to update MTG card: %w", err)
				}
			}
		}

		return nil
	})

	return len(cardsToCreate), len(cardsToUpdate), err
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
