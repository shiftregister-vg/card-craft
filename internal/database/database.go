package database

import (
	"database/sql"
	"fmt"

	"github.com/shiftregister-vg/card-craft/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	*sql.DB
}

// New creates a new database connection using the provided configuration
func New(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &Database{db}, nil
}

func (db *Database) Close() error {
	return db.DB.Close()
}
