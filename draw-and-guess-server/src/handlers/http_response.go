package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPErrorResponse represents an HTTP error response
type HTTPErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSONSuccess sends HTTP success response without envelope (직접 데이터 반환)
func JSONSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// JSONError sends HTTP error response with code and message
func JSONError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, HTTPErrorResponse{
		Code:    code,
		Message: message,
	})
}

// JSONBadRequest sends a 400 Bad Request response
func JSONBadRequest(c *gin.Context, message string) {
	JSONError(c, http.StatusBadRequest, "INVALID_ARGUMENT", message)
}

// JSONNotFound sends a 404 Not Found response
func JSONNotFound(c *gin.Context, message string) {
	JSONError(c, http.StatusNotFound, "NOT_FOUND", message)
}

// JSONInternalError sends a 500 Internal Server Error response
func JSONInternalError(c *gin.Context, message string) {
	JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// JSONUnauthorized sends a 401 Unauthorized response
func JSONUnauthorized(c *gin.Context, message string) {
	JSONError(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// JSONConflict sends a 409 Conflict response
func JSONConflict(c *gin.Context, message string) {
	JSONError(c, http.StatusConflict, "CONFLICT", message)
}
