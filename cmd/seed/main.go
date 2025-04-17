package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"github.com/shiftregister-vg/card-craft/internal/seed"
)

func buildDatabaseURL() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func main() {
	// Parse command line flags
	clear := flag.Bool("clear", false, "Clear the database before seeding")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(".env.localhost", ".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Build database connection string
	connStr := buildDatabaseURL()
	if connStr == "" {
		log.Fatal("Failed to build database connection string from environment variables")
	}

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create seeder
	seeder := seed.NewSeeder(db)

	// Create context
	ctx := context.Background()

	// Clear database if requested
	if *clear {
		fmt.Println("Clearing database...")
		if err := seeder.Clear(ctx); err != nil {
			log.Fatalf("Failed to clear database: %v", err)
		}
		fmt.Println("Database cleared successfully")
	}

	// Seed the database
	fmt.Println("Seeding database...")
	if err := seeder.Seed(ctx); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			log.Fatalf("Database error: %v (Code: %s)", pqErr.Message, pqErr.Code)
		}
		log.Fatalf("Failed to seed database: %v", err)
	}

	fmt.Println("Database seeded successfully")
}
