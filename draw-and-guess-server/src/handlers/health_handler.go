package handlers

import (
	"draw-and-guess-server/src/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.ApiResult{
		Status:  true,
		Message: "Server is healthy",
		Result:  nil,
	})
}
