package config

import (
	"fmt"
	"os"
)

// Config holds everything the app reads from the environment.
// Connection secrets are never hardcoded or committed.
type Config struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	Port          string
	StaticDir     string
}

// Load reads configuration from environment variables and validates
// that the required connection details are present.
func Load() (Config, error) {
	cfg := Config{
		Neo4jURI:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     getenv("NEO4J_USER", "cognodb"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
		Port:          getenv("PORT", "8080"),
		StaticDir:     getenv("STATIC_DIR", "frontend/dist"),
	}
	if cfg.Neo4jURI == "" {
		return cfg, fmt.Errorf("NEO4J_URI is not set (e.g. bolt+s://<instance-id>.databases.cognodb.cloud)")
	}
	if cfg.Neo4jPassword == "" {
		return cfg, fmt.Errorf("NEO4J_PASSWORD is not set")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
