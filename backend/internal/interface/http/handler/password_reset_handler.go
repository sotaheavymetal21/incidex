package handler

import (
	"incidex/internal/interface/http/validator"
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PasswordResetHandler はパスワードリセット関連の HTTP handler を提供します
type PasswordResetHandler struct {
	passwordResetUsecase usecase.PasswordResetUsecase
}

// NewPasswordResetHandler は新しい PasswordResetHandler を作成します
func NewPasswordResetHandler(passwordResetUsecase usecase.PasswordResetUsecase) *PasswordResetHandler {
	return &PasswordResetHandler{passwordResetUsecase: passwordResetUsecase}
}

// RequestPasswordResetRequest はパスワードリセット要求の request body を表します
type RequestPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// RequestPasswordReset はパスワードリセット要求を処理します
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

	// メール列挙攻撃を防ぐために常に成功を返します
	c.JSON(http.StatusOK, gin.H{
		"message": "パスワードリセットのメールを送信しました。メールをご確認ください。",
	})
}

// ResetPasswordRequest はパスワードリセット実行の request body を表します
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=12"`
}

// ResetPassword は実際のパスワードリセットを処理します
func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// パスワード強度をバリデーション（厳格モード）
	if err := validator.ValidatePassword(req.NewPassword, true); err != nil {
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

// ValidateToken はパスワードリセット token が有効かどうかをバリデーションします
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
