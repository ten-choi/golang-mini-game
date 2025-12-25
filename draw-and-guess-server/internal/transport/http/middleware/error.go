package middleware

import (
	"draw-and-guess-server/internal/platform/apperr"
	"draw-and-guess-server/internal/platform/logger"
	"draw-and-guess-server/internal/transport/rest"

	"github.com/gin-gonic/gin"
)

// ErrorHandler handles errors and returns standardized responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			rest.ErrorResponseJSON(c, err)
		}
	}
}

// Recovery recovers from panics and returns a 500 error
func Recovery() gin.HandlerFunc {
	log := logger.GetLogger()

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Panic recovered: %v", err)
				rest.ErrorResponseJSON(c, apperr.NewInternalError("Internal server error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}
