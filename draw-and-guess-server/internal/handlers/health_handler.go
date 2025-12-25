package handlers

import (
	"draw-and-guess-server/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck은 서버 상태를 확인하는 API 핸들러
// GET /health
// 용도: 서버가 정상적으로 동작하는지 확인 (헬스체크)
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.ApiResult{
		Status:  true,
		Message: "Server is healthy",
		Result:  nil,
	})
}
