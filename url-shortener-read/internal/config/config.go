package config

import (
	"fmt"
	"os"
)

type Config struct {
	PORT      string
	DSN       string
	MongoURI  string
	MongoDB   string
	MongoColl string
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
		MongoURI:  os.Getenv("MONGO_URI"),
		MongoDB:   mongoDB,
		MongoColl: mongoColl,
	}, nil
}
