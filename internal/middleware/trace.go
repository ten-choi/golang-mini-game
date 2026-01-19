package middleware

import (
	"draw-and-guess-server/pkg/utils"

	"github.com/gin-gonic/gin"
)

// TraceID middleware adds a unique trace ID to each request
// The trace ID can be used for log correlation and debugging
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID exists in header (for distributed tracing)
		traceID := c.GetHeader("X-Trace-ID")

		// Generate new trace ID if not provided
		if traceID == "" {
			traceID = utils.GenerateTraceID()
		}

		// Store in context for use in handlers and logs
		c.Set("trace_id", traceID)

		// Add to response header for client visibility
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}
