package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/shiftregister-vg/card-craft/internal/config"
	"github.com/shiftregister-vg/card-craft/internal/database"
	"github.com/shiftregister-vg/card-craft/internal/graph"
	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/shiftregister-vg/card-craft/internal/middleware"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationsDir := filepath.Join("migrations")
	if err := db.Migrate(migrationsDir); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)
	rateLimitMiddleware, err := middleware.NewRateLimitMiddleware(cfg.RateLimit, cfg.RateLimitPeriod.String())
	if err != nil {
		log.Fatalf("Error creating rate limiter: %v", err)
	}

	// Create GraphQL server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: graph.NewResolver(nil, nil, nil)}))

	// Create HTTP server with middleware
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", authMiddleware.Middleware(rateLimitMiddleware.Middleware(srv)))

	// Start server
	log.Printf("Server running on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
