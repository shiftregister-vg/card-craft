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

func main() {
	// Parse command line flags
	clear := flag.Bool("clear", false, "Clear the database before seeding")
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
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
