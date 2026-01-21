package handler

import (
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	passwordResetUsecase usecase.PasswordResetUsecase
}

func NewPasswordResetHandler(passwordResetUsecase usecase.PasswordResetUsecase) *PasswordResetHandler {
	return &PasswordResetHandler{passwordResetUsecase: passwordResetUsecase}
}

type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestPasswordReset handles the password reset request
func (h *PasswordResetHandler) RequestPasswordReset(c *gin.Context) {
	var req RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	err := h.passwordResetUsecase.RequestPasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Always return success to prevent email enumeration attacks
	c.JSON(http.StatusOK, gin.H{
		"message": "パスワードリセットのメールを送信しました。メールをご確認ください。",
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ResetPassword handles the actual password reset
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	err := h.passwordResetUsecase.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "パスワードが正常にリセットされました。",
	})
}

// ValidateToken validates if a password reset token is valid
func (h *PasswordResetHandler) ValidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "トークンが必要です",
		})
		return
	}

	valid, err := h.passwordResetUsecase.ValidateToken(c.Request.Context(), token)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": valid,
	})
}
