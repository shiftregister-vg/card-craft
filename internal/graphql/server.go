package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/shiftregister-vg/card-craft/internal/graph"
	"github.com/shiftregister-vg/card-craft/internal/graph/generated"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

// NewServer creates a new GraphQL server with the given resolver
func NewServer(resolver *graph.Resolver) http.Handler {
	// Create the GraphQL server
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: resolver}))

	// Configure the server
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Configure query caching
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Enable introspection
	srv.Use(extension.Introspection{})

	// Enable automatic persisted queries
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Create CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept"},
		AllowCredentials: true,
		ExposedHeaders:   []string{"Set-Cookie"},
		Debug:            true,
	})

	// Create the playground handler
	playgroundHandler := playground.Handler("GraphQL playground", "/query")

	// Create the main handler
	mux := http.NewServeMux()
	mux.Handle("/", playgroundHandler)
	mux.Handle("/query", c.Handler(srv))

	return mux
}

// NewContextMiddleware creates a middleware that adds the resolver to the request context
func NewContextMiddleware(resolver *graph.Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "resolver", resolver)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
