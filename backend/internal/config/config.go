// Package config loads runtime configuration from the environment.
package config

import (
	"os"
	"strings"
)

type Config struct {
	Addr        string
	DatabaseURL string
	LogLevel    string
	// SeedIfEmpty populates a demo flow the first time the database is empty.
	SeedIfEmpty bool
}

func Load() Config {
	return Config{
		Addr:        env("SERVER_ADDR", ":8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://tracker:tracker@127.0.0.1:5432/tracker?sslmode=disable"),
		LogLevel:    env("LOG_LEVEL", "info"),
		SeedIfEmpty: env("SEED_IF_EMPTY", "true") == "true",
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
