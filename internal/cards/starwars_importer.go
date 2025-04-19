package cards

import (
	"context"
)

// StarWarsImporter handles importing Star Wars: Unlimited card data
type StarWarsImporter struct {
	*BaseImporter
}

// NewStarWarsImporter creates a new Star Wars: Unlimited importer
func NewStarWarsImporter(store *CardStore) *StarWarsImporter {
	return &StarWarsImporter{
		BaseImporter: NewBaseImporter(store, "starwars"),
	}
}

// Import implements the Importer interface
func (i *StarWarsImporter) Import(ctx context.Context, store *CardStore) error {
	return i.ImportLatestSets(ctx)
}

// ImportSet imports all cards from a specific set
func (i *StarWarsImporter) ImportSet(ctx context.Context, setID string) error {
	// TODO: Implement Star Wars: Unlimited set import
	// This will need to be implemented once we have access to the Star Wars: Unlimited API or data source
	return nil
}

// ImportLatestSets imports cards from the latest Star Wars: Unlimited sets
func (i *StarWarsImporter) ImportLatestSets(ctx context.Context) error {
	// TODO: Implement Star Wars: Unlimited latest sets import
	// This will need to be implemented once we have access to the Star Wars: Unlimited API or data source
	return nil
}
