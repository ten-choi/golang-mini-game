package rest

import (
	"net/http"

	"draw-and-guess-server/internal/platform/apperr"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents error response structure
// Used for REST endpoints only (not GraphQL)
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    int               `json:"code"`               // HTTP status code
	Message string            `json:"message"`            // Human-readable error message
	Details map[string]string `json:"details,omitempty"`  // Field-specific validation errors
	TraceID string            `json:"trace_id,omitempty"` // Request ID for log tracing
}

// SuccessResponse sends data directly (unenveloped)
// REST ONLY: For REST endpoints like file upload, OAuth login, webhooks
// GraphQL: Do NOT use this - return data directly in resolvers
// HTTP status code indicates success (200), no envelope needed
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// ErrorResponseJSON sends a standardized error response
// REST ONLY: For REST endpoints only
// GraphQL: gqlgen handles errors automatically
func ErrorResponseJSON(c *gin.Context, err error) {
	appErr := apperr.GetAppError(err)

	// Get trace ID from context if available
	traceID := ""
	if id, exists := c.Get("trace_id"); exists {
		if idStr, ok := id.(string); ok {
			traceID = idStr
		}
	}

	c.JSON(appErr.Code, ErrorResponse{
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			TraceID: traceID,
		},
	})
}

// ErrorResponseWithDetails sends error response with field-specific details
// REST ONLY: Useful for validation errors
// Example: {"error": {"code": 400, "message": "Validation failed", "details": {"email": "invalid format"}}}
func ErrorResponseWithDetails(c *gin.Context, err error, details map[string]string) {
	appErr := apperr.GetAppError(err)

	traceID := ""
	if id, exists := c.Get("trace_id"); exists {
		if idStr, ok := id.(string); ok {
			traceID = idStr
		}
	}

	c.JSON(appErr.Code, ErrorResponse{
		Error: ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: details,
			TraceID: traceID,
		},
	})
}

// CreatedResponse sends a 201 Created response with data
// REST ONLY: Use for resource creation endpoints
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// NoContentResponse sends a 204 No Content response
// REST ONLY: Use for successful DELETE or UPDATE operations with no body
func NoContentResponse(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SendJSON is an alias for SuccessResponse with more verb-like naming
// Some teams prefer verb-like function names
func SendJSON(c *gin.Context, data interface{}) {
	SuccessResponse(c, data)
}
