package main

import (
	"context"
	"flag"
	"log"

	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/config"
	"github.com/shiftregister-vg/card-craft/internal/database"
)

func main() {
	// Parse command line flags
	gameType := flag.String("game", "", "Game type to import (pokemon, lorcana, starwars)")
	flag.Parse()

	if *gameType == "" {
		log.Fatal("Game type is required")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create card store
	cardStore := cards.NewCardStore(db.DB)

	// Create game-specific stores
	pokemonStore := cards.NewPokemonCardStore(db.DB)

	// Select importer based on game type
	var importer cards.Importer
	switch *gameType {
	case "pokemon":
		importer = cards.NewPokemonImporter(cardStore, pokemonStore)
	case "lorcana":
		importer = cards.NewLorcanaImporter(cardStore)
	case "starwars":
		importer = cards.NewStarWarsImporter(cardStore)
	default:
		log.Fatalf("Unsupported game type: %s", *gameType)
	}

	// Run import
	if err := importer.Import(context.Background(), cardStore); err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Printf("Import completed successfully")
}
