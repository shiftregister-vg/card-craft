package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/config"
	"github.com/shiftregister-vg/card-craft/internal/database"
	"github.com/shiftregister-vg/card-craft/internal/graph"
	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/shiftregister-vg/card-craft/internal/middleware"
	"github.com/shiftregister-vg/card-craft/internal/models"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
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

	// Initialize stores
	cardStore := models.NewCardStore(db.DB)
	deckStore := models.NewDeckStore(db.DB)
	userStore := models.NewUserStore(db.DB)

	// Initialize services
	authService := auth.NewService(cfg.JWTSecret, userStore)
	searchService := cards.NewSearchService(cardStore)

	// Create GraphQL server
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: graph.NewResolver(authService, cardStore, deckStore, searchService),
	})

	// Create GraphQL handler with recommended configuration
	graphqlHandler := handler.New(schema)
	graphqlHandler.AddTransport(transport.Options{})
	graphqlHandler.AddTransport(transport.GET{})
	graphqlHandler.AddTransport(transport.POST{})
	graphqlHandler.AddTransport(transport.MultipartForm{})

	// Configure query caching
	graphqlHandler.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Enable introspection
	graphqlHandler.Use(extension.Introspection{})

	// Enable automatic persisted queries
	graphqlHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Create HTTP server with middleware
	mux := http.NewServeMux()
	mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	mux.Handle("/query", authMiddleware.Middleware(rateLimitMiddleware.Middleware(graphqlHandler)))

	// Start server
	log.Printf("Server running on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
