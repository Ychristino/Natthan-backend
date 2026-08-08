package config

import (
	"log"
	"os"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  string // ex: "15m"
	JWTRefreshTokenTTL string // ex: "168h" (7 dias)

	// App
	Env string
}

func Load() *Config {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTAccessTokenTTL:  getEnv("JWT_ACCESS_TOKEN_TTL", "15m"),
		JWTRefreshTokenTTL: getEnv("JWT_REFRESH_TOKEN_TTL", "168h"),
		Env:                getEnv("ENV", "development"),
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL é obrigatório")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET é obrigatório")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
