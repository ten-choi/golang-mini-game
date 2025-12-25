package middleware

import (
	"draw-and-guess-server/internal/common"

	"github.com/gin-gonic/gin"
)

// ErrorHandler handles errors and returns standardized responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			common.ErrorResponse(c, err)
		}
	}
}

// Recovery recovers from panics and returns a 500 error
func Recovery() gin.HandlerFunc {
	logger := common.GetLogger()
	
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered: %v", err)
				common.ErrorResponse(c, common.NewInternalError("Internal server error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}
