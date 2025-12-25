package middleware

import (
	"time"

	"draw-and-guess-server/internal/common"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs all incoming HTTP requests
func RequestLogger() gin.HandlerFunc {
	logger := common.GetLogger()
	
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logger.Info("HTTP %s %s | Status: %d | Latency: %v | IP: %s",
			method, path, statusCode, latency, clientIP)
	}
}
