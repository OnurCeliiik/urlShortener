package main

import (
	"OnurCeliiik/urlShortener-write/database"
	"OnurCeliiik/urlShortener-write/internal/config"
	"OnurCeliiik/urlShortener-write/internal/handler"
	"OnurCeliiik/urlShortener-write/internal/repository"
	"OnurCeliiik/urlShortener-write/internal/service"
	"OnurCeliiik/urlShortener-write/migrations"
	"OnurCeliiik/urlShortener-write/routes"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := database.NewDB(cfg.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db, migrations.SQL); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	repo := repository.NewShortenRepository(db)
	svc := service.NewShortenService(repo, cfg.BaseURL)
	h := handler.NewShortenHandler(svc)

	router := gin.Default()
	routes.SetupRoutes(router, &routes.Dependencies{
		ShortenHandler: h,
		DB:             db,
	})

	server := &http.Server{
		Addr:              ":" + cfg.PORT,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("write service listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start writer server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
}
