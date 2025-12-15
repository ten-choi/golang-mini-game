package handlers

import (
	"draw-and-guess-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError standardizes error responses for handlers.
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, models.ApiResult{
		Status:  false,
		Message: message,
		Result:  nil,
	})
}

// respondSuccess standardizes success responses for handlers.
func respondSuccess(c *gin.Context, message string, result interface{}) {
	c.JSON(http.StatusOK, models.ApiResult{
		Status:  true,
		Message: message,
		Result:  result,
	})
}
