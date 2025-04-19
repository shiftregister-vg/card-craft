package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	JWTExpiration     time.Duration
	RateLimit         int
	RateLimitPeriod   time.Duration
	Port              string
	Environment       string
	EnableCardImports bool
}

func Load() (*Config, error) {
	projectRoot := os.Getenv("DEVBOX_PROJECT_ROOT")
	if projectRoot == "" {
		return nil, fmt.Errorf("DEVBOX_PROJECT_ROOT environment variable is not set")
	}

	// First, load the .env file from DEVBOX_PROJECT_ROOT
	envPath := filepath.Join(projectRoot, ".env")
	if err := godotenv.Load(envPath); err != nil {
		// Don't return error if .env doesn't exist
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Then, load .env.localhost from DEVBOX_PROJECT_ROOT
	envLocalhostPath := filepath.Join(projectRoot, ".env.localhost")
	if err := godotenv.Load(envLocalhostPath); err != nil {
		return nil, fmt.Errorf("failed to load environment file %s: %w", envLocalhostPath, err)
	}

	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT", "100"))
	jwtExpiration, _ := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	rateLimitPeriod, _ := time.ParseDuration(getEnv("RATE_LIMIT_PERIOD", "1m"))
	enableCardImports, _ := strconv.ParseBool(getEnv("ENABLE_CARD_IMPORTS", "false"))

	return &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            port,
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "cardcraft"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key-here"),
		JWTExpiration:     jwtExpiration,
		RateLimit:         rateLimit,
		RateLimitPeriod:   rateLimitPeriod,
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		EnableCardImports: enableCardImports,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
