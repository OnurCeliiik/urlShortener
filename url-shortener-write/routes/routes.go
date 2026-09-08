package routes

import (
	"OnurCeliiik/urlShortener-write/internal/handler"
	"OnurCeliiik/urlShortener-write/internal/obs"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HealthChecker interface {
	Ping() error
}

type Dependencies struct {
	ShortenHandler *handler.ShortenHandler
	DB             HealthChecker
	Logger         *slog.Logger
}

func SetupRoutes(router *gin.Engine, deps *Dependencies) {
	router.Use(obs.RequestID())
	if deps.Logger != nil {
		router.Use(obs.AccessLog(deps.Logger))
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5500",
			"http://127.0.0.1:5500",
		},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "X-Request-ID"},
		ExposeHeaders: []string{"X-Request-ID"},
	}))

	router.GET("/health", func(c *gin.Context) {
		if deps.DB != nil {
			if err := deps.DB.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"service": "write",
					"status":  "unhealthy",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service": "write",
			"status":  "ok",
		})
	})

	router.POST("/shorten", deps.ShortenHandler.Shorten)
}
