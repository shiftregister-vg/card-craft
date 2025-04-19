package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/cards"
	"github.com/shiftregister-vg/card-craft/internal/config"
	"github.com/shiftregister-vg/card-craft/internal/database"
	"github.com/shiftregister-vg/card-craft/internal/graph"
	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/shiftregister-vg/card-craft/internal/middleware"
	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/shiftregister-vg/card-craft/internal/scheduler"
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

	// Initialize stores
	cardStore := cards.NewCardStore(db.DB)
	pokemonStore := cards.NewPokemonCardStore(db.DB)
	deckStore := models.NewDeckStore(db.DB)
	userStore := models.NewUserStore(db.DB)
	collectionStore := models.NewCollectionStore(db.DB)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret, userStore)
	rateLimitMiddleware, err := middleware.NewRateLimitMiddleware(cfg.RateLimit, cfg.RateLimitPeriod.String())
	if err != nil {
		log.Fatalf("Error creating rate limiter: %v", err)
	}

	// Initialize services
	authService := auth.NewService(cfg.JWTSecret, userStore)

	// Initialize and start the scheduler for card imports
	pokemonImporter := cards.NewPokemonImporter(cardStore, pokemonStore)
	lorcanaImporter := cards.NewLorcanaImporter(cardStore)
	starWarsImporter := cards.NewStarWarsImporter(cardStore)

	sched := scheduler.NewScheduler(pokemonImporter, lorcanaImporter, starWarsImporter)
	sched.Start()
	defer sched.Stop()

	// Create GraphQL server
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: graph.NewResolver(
			db.DB,
			cardStore,
			pokemonStore,
			deckStore,
			collectionStore,
			authService,
		),
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
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", rateLimitMiddleware.Middleware(authMiddleware.Middleware(graphqlHandler)))

	// Start server
	log.Printf("Server is running on http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
