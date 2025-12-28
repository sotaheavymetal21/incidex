package handler

import (
	"incidex/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents the JSON structure for error responses
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HandleError processes errors and returns appropriate HTTP responses
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if it's a DomainError
	if domainErr, ok := domain.AsDomainError(err); ok {
		response := ErrorResponse{
			Error:   string(domainErr.Code),
			Message: domainErr.Message,
		}

		// Include details if present
		if len(domainErr.Details) > 0 {
			response.Details = domainErr.Details
		}

		c.JSON(domainErr.StatusCode, response)
		return
	}

	// Default error response for unknown errors
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   string(domain.ErrCodeInternal),
		Message: "An internal error occurred",
	})
}

// HandleValidationError handles request validation errors
func HandleValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   string(domain.ErrCodeValidation),
		Message: "Invalid request data",
		Details: map[string]interface{}{
			"validation_error": err.Error(),
		},
	})
}
