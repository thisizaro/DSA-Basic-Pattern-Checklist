// Package config centralizes all environment-based configuration.
// Nothing else in the codebase should call os.Getenv directly — this
// keeps configuration discoverable in one place and easy to extend.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds every configurable value the app needs at startup.
type Config struct {
	// Port the HTTP server listens on.
	Port string

	// DatabaseURL is a full Postgres connection string.
	// Works with Supabase, Neon, or any standard Postgres provider.
	// Example: postgres://user:pass@host:5432/dbname?sslmode=require
	DatabaseURL string

	// JWTSecret signs and verifies auth tokens. Must be long and random in production.
	JWTSecret string

	// JWTExpiry controls how long an access token stays valid.
	JWTExpiry time.Duration

	// Environment is "development" or "production". Affects cookie security flags.
	Environment string

	// GoogleClientID / GoogleClientSecret are the OAuth2 web client
	// credentials from Google Cloud Console (APIs & Services > Credentials).
	GoogleClientID     string
	GoogleClientSecret string

	// GoogleRedirectURL must exactly match an "Authorized redirect URI"
	// configured for the OAuth client, e.g.
	// http://localhost:8080/api/auth/google/callback in dev, or
	// https://yourdomain.com/api/auth/google/callback in production.
	GoogleRedirectURL string

	// FrontendBaseURL is where the browser gets redirected after OAuth
	// completes. Leave empty for the default same-process setup (server
	// serves both API and static frontend); set to a full origin (e.g.
	// https://your-frontend.vercel.app) only if the frontend is hosted
	// separately from this backend.
	FrontendBaseURL string
}

// Load reads configuration from a .env file (if present) and the process
// environment, validates required fields, and returns a ready-to-use Config.
func Load() (*Config, error) {
	// Best-effort load of a local .env file. Ignored if it doesn't exist —
	// in production, env vars are usually injected by the platform instead.
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnvOrDefault("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		Environment:        getEnvOrDefault("ENVIRONMENT", "development"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		FrontendBaseURL:    os.Getenv("FRONTEND_BASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (postgres connection string)")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required (use a long random string)")
	}
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required (see README for setup)")
	}
	if cfg.GoogleRedirectURL == "" {
		return nil, fmt.Errorf("GOOGLE_REDIRECT_URL is required, e.g. http://localhost:8080/api/auth/google/callback")
	}

	cfg.JWTExpiry = 7 * 24 * time.Hour // 7 days

	return cfg, nil
}

// IsProduction reports whether the app is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
