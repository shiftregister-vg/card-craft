package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/models"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserContextKey is the key used to store the user in the context
	UserContextKey ContextKey = "user"
)

// Middleware handles authentication for HTTP requests
func Middleware(service *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check the Authorization header format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate the token
			userID, err := service.ValidateToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Parse the user ID
			uuid, err := uuid.Parse(userID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Get the user from the database
			user, err := service.UserStore.FindByID(uuid)
			if err != nil || user == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Add the user to the context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext returns the user from the context
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

// RequireAuth is a middleware that requires authentication
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
