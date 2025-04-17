package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

type Migration struct {
	Version string
	Up      string
	Down    string
}

func (db *Database) Migrate(migrationsDir string) error {
	// Create migrations table if it doesn't exist
	if err := db.createMigrationsTable(); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get all migration files
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("error reading migration files: %w", err)
	}

	// Sort files by name
	sort.Strings(files)

	// Get applied migrations
	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".sql")
		if _, ok := applied[version]; !ok {
			content, err := fs.ReadFile(os.DirFS(migrationsDir), filepath.Base(file))
			if err != nil {
				return fmt.Errorf("error reading migration file %s: %w", file, err)
			}

			if err := db.applyMigration(version, string(content)); err != nil {
				return fmt.Errorf("error applying migration %s: %w", version, err)
			}
		}
	}

	return nil
}

func (db *Database) createMigrationsTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (db *Database) getAppliedMigrations() (map[string]struct{}, error) {
	rows, err := db.Query("SELECT version FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = struct{}{}
	}

	return applied, nil
}

func (db *Database) applyMigration(version, sql string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sql); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO migrations (version) VALUES ($1)", version); err != nil {
		return err
	}

	return tx.Commit()
}
