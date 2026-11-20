package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr          string
	MongoURI          string
	MongoDatabase     string
	RedisURL          string
	JWTSecret         string
	JWTIssuer         string
	JWTExpiry         time.Duration
	AllowedOrigins    []string
	CookieSecure      bool
	MaxBodyBytes      int64
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:        envOrDefault("HTTP_ADDR", ":8080"),
		MongoURI:        envOrDefault("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:   envOrDefault("MONGO_DATABASE", "live_polling"),
		RedisURL:        envOrDefault("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		JWTIssuer:       envOrDefault("JWT_ISSUER", "live-polling-api"),
		JWTExpiry:       durationOrDefault("JWT_EXPIRY", 24*time.Hour),
		AllowedOrigins:  splitCSV(envOrDefault("ALLOWED_ORIGINS", "http://localhost:5173")),
		CookieSecure:    boolOrDefault("COOKIE_SECURE", false),
		MaxBodyBytes:    int64(intOrDefault("MAX_BODY_BYTES", 1<<20)),
		ShutdownTimeout: durationOrDefault("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if cfg.MaxBodyBytes <= 0 {
		return Config{}, errors.New("MAX_BODY_BYTES must be greater than zero")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func intOrDefault(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func boolOrDefault(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}
