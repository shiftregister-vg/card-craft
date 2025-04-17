package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shiftregister-vg/card-craft/internal/auth"
)

// NewRouter creates a new router with all routes configured
func NewRouter(authService *auth.Service) *mux.Router {
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/api/auth/register", authService.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/auth/login", authService.LoginHandler).Methods("POST")

	// Protected routes
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(auth.Middleware(authService))
	protected.HandleFunc("/auth/refresh", auth.RefreshTokenHandler(authService)).Methods("POST")

	// GraphQL routes
	r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// GraphQL handler will be added here
	}).Methods("GET", "POST")

	return r
}
