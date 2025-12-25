package rest

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthCheck checks if the server is healthy
// GET /health
// Used for: Server health monitoring
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
	})
}
