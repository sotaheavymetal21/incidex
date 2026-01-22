package handler

import (
	"incidex/internal/interface/http/validator"
	"incidex/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

type RegisterRequest struct {
	Name           string `json:"name" binding:"required,max=50"`
	Email          string `json:"email" binding:"required,email,max=254"`
	Password       string `json:"password" binding:"required,min=8"`
	EmployeeNumber string `json:"employee_number" binding:"required,max=20"`
	Department     string `json:"department" binding:"required,max=50"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Custom validation
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

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		HandleValidationError(c, err)
		return
	}

	// Custom validation
	if err := validator.ValidateEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authResponse, err := h.authUsecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		HandleError(c, err)
		return
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(
		"refresh_token",           // name
		authResponse.RefreshToken, // value
		7*24*60*60,                // maxAge in seconds (7 days)
		"/",                       // path
		"",                        // domain (empty for current domain)
		false,                     // secure (set to true in production with HTTPS)
		true,                      // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"access_token": authResponse.AccessToken,
		"user":         authResponse.User,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"` // Optional: can be sent in body or cookie
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	// Try to get refresh token from cookie first
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		// Fallback to request body
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

	// Set new refresh token as httpOnly cookie
	c.SetCookie(
		"refresh_token",           // name
		authResponse.RefreshToken, // value
		7*24*60*60,                // maxAge in seconds (7 days)
		"/",                       // path
		"",                        // domain
		false,                     // secure (set to true in production with HTTPS)
		true,                      // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"access_token": authResponse.AccessToken,
		"user":         authResponse.User,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Try to get refresh token from cookie
	refreshToken, _ := c.Cookie("refresh_token")

	if refreshToken != "" {
		// Revoke the refresh token
		if err := h.authUsecase.Logout(c.Request.Context(), refreshToken); err != nil {
			// Log error but don't fail the logout
			// User should still be logged out on the client side
		}
	}

	// Clear the refresh token cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1, // maxAge -1 deletes the cookie
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
