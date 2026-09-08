package routes

import (
	"OnurCeliiik/urlShortener-read/internal/handler"
	"OnurCeliiik/urlShortener-read/internal/obs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthChecker interface {
	Ping() error
}

type Dependencies struct {
	ResolveHandler *handler.ResolveHandler
	DB             HealthChecker
	Logger         *slog.Logger
}

func SetupRoutes(router *gin.Engine, deps *Dependencies) {
	router.Use(obs.RequestID())
	if deps.Logger != nil {
		router.Use(obs.AccessLog(deps.Logger))
	}

	router.GET("/health", func(c *gin.Context) {
		if deps.DB != nil {
			if err := deps.DB.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"service": "read",
					"status":  "unhealthy",
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service": "read",
			"status":  "ok",
		})
	})

	router.GET("/:code", deps.ResolveHandler.Redirect)
}
