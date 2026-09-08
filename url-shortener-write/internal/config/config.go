package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	PORT      string
	DSN       string
	BaseURL   string
	MongoURI  string
	MongoDB   string
	MongoColl string
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

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "urlshortener"
	}

	mongoColl := os.Getenv("MONGO_COLLECTION")
	if mongoColl == "" {
		mongoColl = "events"
	}

	return &Config{
		PORT:      port,
		DSN:       dsn,
		BaseURL:   baseURL,
		MongoURI:  os.Getenv("MONGO_URI"),
		MongoDB:   mongoDB,
		MongoColl: mongoColl,
	}, nil
}
