package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
	JWTExpiration   time.Duration
	RateLimit       int
	RateLimitPeriod time.Duration
	Port            string
	Environment     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT", "100"))
	jwtExpiration, _ := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	rateLimitPeriod, _ := time.ParseDuration(getEnv("RATE_LIMIT_PERIOD", "1m"))

	return &Config{
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          port,
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "cardcraft"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-here"),
		JWTExpiration:   jwtExpiration,
		RateLimit:       rateLimit,
		RateLimitPeriod: rateLimitPeriod,
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
