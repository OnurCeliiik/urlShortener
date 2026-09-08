package config

import (
	"fmt"
	"os"
)

type Config struct {
	PORT string
	DSN  string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dsn := os.Getenv("DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DSN is required")
	}

	return &Config{
		PORT: port,
		DSN:  dsn,
	}, nil
}
