package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/114windd/restapi/internal/logger"
	"github.com/114windd/restapi/internal/metrics"
)

// LoggingMiddleware creates a Gin middleware for request logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log after processing
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		entry := logger.LogRequest(method, path, GetUserIDFromContext(c))
		entry = entry.WithFields(map[string]interface{}{
			"status_code": statusCode,
			"duration_ms": duration.Milliseconds(),
			"client_ip":   c.ClientIP(),
		})

		if statusCode >= 400 {
			entry.Warn("Request completed with error")
		} else {
			entry.Info("Request completed successfully")
		}
	}
}

// PrometheusMiddleware creates a Gin middleware for Prometheus metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()

		metrics.RecordHTTPRequest(method, path, statusCode, duration)
	}
}
