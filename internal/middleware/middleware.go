package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shiftregister-vg/card-craft/internal/auth"
	"github.com/shiftregister-vg/card-craft/internal/models"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type AuthMiddleware struct {
	jwtSecret string
	userStore *models.UserStore
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func NewAuthMiddleware(jwtSecret string, userStore *models.UserStore) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret, userStore: userStore}
}

func (m *AuthMiddleware) isAuthenticationRequired(r *http.Request) bool {
	if r.Method != "POST" {
		return false
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return true
	}
	// Reset the body for subsequent reads
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	var req graphqlRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return true
	}

	// Allow unauthenticated access to login and register mutations
	if strings.Contains(req.Query, "mutation Login") || strings.Contains(req.Query, "mutation Register") {
		return false
	}

	return true
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.isAuthenticationRequired(r) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		uuid, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		user, err := m.userStore.FindByID(uuid)
		if err != nil || user == nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type RateLimitMiddleware struct {
	limiter *limiter.Limiter
}

func NewRateLimitMiddleware(rate int, period string) (*RateLimitMiddleware, error) {
	duration, err := time.ParseDuration(period)
	if err != nil {
		return nil, err
	}

	rateLimit := limiter.Rate{
		Period: duration,
		Limit:  int64(rate),
	}

	store := memory.NewStore()
	limiter := limiter.New(store, rateLimit)

	return &RateLimitMiddleware{limiter: limiter}, nil
}

func (m *RateLimitMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context, err := m.limiter.Get(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if context.Reached {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
