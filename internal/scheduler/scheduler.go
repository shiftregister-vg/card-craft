package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/shiftregister-vg/card-craft/internal/cards"
)

// Scheduler handles periodic tasks for the application
type Scheduler struct {
	importers []cards.CardImporter
	store     *cards.CardStore
	stop      chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(store *cards.CardStore, importers ...cards.CardImporter) *Scheduler {
	return &Scheduler{
		importers: importers,
		store:     store,
		stop:      make(chan struct{}),
	}
}

// Start begins running scheduled tasks
func (s *Scheduler) Start() {
	go s.runCardImports()
}

// Stop stops all scheduled tasks
func (s *Scheduler) Stop() {
	close(s.stop)
}

// runCardImports runs the card import process daily
func (s *Scheduler) runCardImports() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run immediately on start
	s.importCards()

	for {
		select {
		case <-ticker.C:
			s.importCards()
		case <-s.stop:
			return
		}
	}
}

// importCards handles the card import process for all games
func (s *Scheduler) importCards() {
	for _, importer := range s.importers {
		log.Printf("Starting scheduled card import for %s...", importer.GetGame())
		if err := importer.Import(context.Background(), s.store); err != nil {
			log.Printf("Error importing %s cards: %v", importer.GetGame(), err)
			continue
		}
		log.Printf("Scheduled %s card import completed successfully", importer.GetGame())
	}
}
