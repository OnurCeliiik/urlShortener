package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PORT    string
	DSN     string
	BaseURL string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:" + port
	}

	return &Config{
		PORT:    port,
		DSN:     dsn,
		BaseURL: baseURL,
	}, nil
}
