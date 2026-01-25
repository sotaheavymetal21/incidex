package handler

import (
	"incidex/internal/interface/http/validator"
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler は認証関連の HTTP handler を提供します
type AuthHandler struct {
	authUsecase  usecase.AuthUsecase
	isProduction bool
}

// NewAuthHandler は新しい AuthHandler を作成します
func NewAuthHandler(authUsecase usecase.AuthUsecase, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authUsecase:  authUsecase,
		isProduction: isProduction,
	}
}

// RegisterRequest はユーザー登録の request body を表します
type RegisterRequest struct {
	Name           string `json:"name" binding:"required,max=50"`
	Email          string `json:"email" binding:"required,email,max=254"`
	Password       string `json:"password" binding:"required,min=8"`
	EmployeeNumber string `json:"employee_number" binding:"required,max=20"`
	Department     string `json:"department" binding:"required,max=50"`
}

// Register は新規ユーザーを登録します
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// カスタムバリデーション
	if err := validator.ValidateName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidatePassword(req.Password, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateEmployeeNumber(req.EmployeeNumber); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateDepartment(req.Department); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authUsecase.Register(c.Request.Context(), req.Name, req.Email, req.Password, req.EmployeeNumber, req.Department)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// LoginRequest はログインの request body を表します
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required"`
}

// Login はユーザーを認証しトークンを発行します
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// カスタムバリデーション
	if err := validator.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		HandleError(c, err)
		return
	}

	// refresh token を httpOnly cookie として設定
	c.SetCookie(
		"refresh_token",           // 名前
		authResponse.RefreshToken, // 値
		7*24*60*60,                // 有効期間（秒）: 7日間
		"/",                       // パス
		"",                        // ドメイン（空文字で現在のドメイン）
		h.isProduction,            // secure（本番環境では HTTPS のため true）
		true,                      // httpOnly
	)
	// CSRF 対策として SameSite 属性を設定
	c.SetSameSite(http.SameSiteStrictMode)

	c.JSON(http.StatusOK, gin.H{
		"access_token": authResponse.AccessToken,
		"user":         authResponse.User,
	})
}

// RefreshRequest はトークン更新の request body を表します
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"` // オプション: body または cookie で送信可能
}

// Refresh はアクセストークンを更新します
func (h *AuthHandler) Refresh(c *gin.Context) {
	// まず cookie から refresh token の取得を試みる
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// request body にフォールバック
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
		return
	}

	authResponse, err := h.authUsecase.RefreshAccessToken(c.Request.Context(), refreshToken)
	if err != nil {
		HandleError(c, err)
		return
	}

	// 新しい refresh token を httpOnly cookie として設定
	c.SetCookie(
		"refresh_token",           // 名前
		authResponse.RefreshToken, // 値
		7*24*60*60,                // 有効期間（秒）: 7日間
		"/",                       // パス
		"",                        // ドメイン
		h.isProduction,            // secure（本番環境では HTTPS のため true）
		true,                      // httpOnly
	)
	// CSRF 対策として SameSite 属性を設定
	c.SetSameSite(http.SameSiteStrictMode)

	c.JSON(http.StatusOK, gin.H{
		"access_token": authResponse.AccessToken,
		"user":         authResponse.User,
	})
}

// Logout はユーザーをログアウトします
func (h *AuthHandler) Logout(c *gin.Context) {
	// cookie から refresh token の取得を試みる
	refreshToken, _ := c.Cookie("refresh_token")

	if refreshToken != "" {
		// refresh token を無効化
		if err := h.authUsecase.Logout(c.Request.Context(), refreshToken); err != nil {
			// error をログに記録するが、ログアウトは失敗させない
			// クライアント側ではユーザーはログアウトされた状態になるべき
		}
	}

	// refresh token cookie をクリア
	c.SetCookie(
		"refresh_token",
		"",
		-1, // maxAge -1 で cookie を削除
		"/",
		"",
		h.isProduction,
		true,
	)
	c.SetSameSite(http.SameSiteStrictMode)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
