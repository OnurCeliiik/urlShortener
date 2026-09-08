package obs

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestIDKey = "request_id"
	StartedAtKey = "started_at"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		c.Set(RequestIDKey, id)
		c.Set(StartedAtKey, time.Now())
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		attrs := []any{
			"request_id", RequestIDFrom(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", ElapsedMS(c),
		}

		if c.Request.URL.Path == "/health" {
			logger.Debug("request", attrs...)
			return
		}
		logger.Info("request", attrs...)
	}
}

func RequestIDFrom(c *gin.Context) string {
	v, ok := c.Get(RequestIDKey)
	if !ok {
		return ""
	}
	id, _ := v.(string)
	return id
}

func ElapsedMS(c *gin.Context) int64 {
	v, ok := c.Get(StartedAtKey)
	if !ok {
		return 0
	}
	started, ok := v.(time.Time)
	if !ok {
		return 0
	}
	return time.Since(started).Milliseconds()
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b)
}
