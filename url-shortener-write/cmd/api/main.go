package main

import (
	"OnurCeliiik/urlShortener-write/database"
	"OnurCeliiik/urlShortener-write/internal/audit"
	"OnurCeliiik/urlShortener-write/internal/config"
	"OnurCeliiik/urlShortener-write/internal/handler"
	"OnurCeliiik/urlShortener-write/internal/obs"
	"OnurCeliiik/urlShortener-write/internal/repository"
	"OnurCeliiik/urlShortener-write/internal/service"
	"OnurCeliiik/urlShortener-write/migrations"
	"OnurCeliiik/urlShortener-write/routes"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	logger := obs.NewLogger("write")
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		logger.Info("no .env file found, using environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	db, err := database.NewDB(cfg.DSN)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db, migrations.SQL); err != nil {
		logger.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	auditor := audit.Noop()
	if cfg.MongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		connected, err := audit.Connect(ctx, cfg.MongoURI, cfg.MongoDB, cfg.MongoColl, logger)
		cancel()
		if err != nil {
			logger.Warn("audit store unavailable, continuing without Mongo events", "err", err)
		} else {
			auditor = connected
			logger.Info("audit store connected", "db", cfg.MongoDB, "collection", cfg.MongoColl)
		}
	}
	defer auditor.Close()

	repo := repository.NewShortenRepository(db)
	svc := service.NewShortenService(repo, cfg.BaseURL)
	h := handler.NewShortenHandler(svc, auditor)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	routes.SetupRoutes(router, &routes.Dependencies{
		ShortenHandler: h,
		DB:             db,
		Logger:         logger,
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
		logger.Info("listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start writer server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown", "err", err)
		os.Exit(1)
	}
}
