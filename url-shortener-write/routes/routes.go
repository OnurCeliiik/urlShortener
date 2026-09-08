package routes

import (
	"OnurCeliiik/urlShortener-write/internal/handler"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthChecker interface {
	Ping() error
}

type Dependencies struct {
	ShortenHandler *handler.ShortenHandler
	DB             HealthChecker
}

func SetupRoutes(router *gin.Engine, deps *Dependencies) {
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
