package handler

import (
	"incidex/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse は error response の JSON 構造を表します
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HandleError は error を処理し適切な HTTP response を返します
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// DomainError かどうかをチェック
	if domainErr, ok := domain.AsDomainError(err); ok {
		response := ErrorResponse{
			Error:   string(domainErr.Code),
			Message: domainErr.Message,
		}

		// 詳細がある場合は含める
		if len(domainErr.Details) > 0 {
			response.Details = domainErr.Details
		}

		c.JSON(domainErr.StatusCode, response)
		return
	}

	// 不明な error に対するデフォルトの error response
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   string(domain.ErrCodeInternal),
		Message: "An internal error occurred",
	})
}

// HandleValidationError は request バリデーション error を処理します
func HandleValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   string(domain.ErrCodeValidation),
		Message: "Invalid request data",
		Details: map[string]interface{}{
			"validation_error": err.Error(),
		},
	})
}
