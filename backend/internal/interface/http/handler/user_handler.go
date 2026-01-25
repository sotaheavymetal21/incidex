package handler

import (
	"incidex/internal/domain"
	"incidex/internal/interface/http/validator"
	"incidex/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler はユーザー関連の HTTP handler を提供します
type UserHandler struct {
	userUsecase usecase.UserUsecase
}

// NewUserHandler は新しい UserHandler を作成します
func NewUserHandler(u usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: u}
}

// GetAll はすべてのユーザーを取得します
func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.userUsecase.GetAllUsers(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetByID は指定された ID のユーザーを取得します
func (h *UserHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUserRequest はユーザー作成の request body を表します
type CreateUserRequest struct {
	Email          string      `json:"email" binding:"required,email,max=254"`
	Password       string      `json:"password" binding:"required,min=6"`
	Name           string      `json:"name" binding:"required,max=50"`
	Role           domain.Role `json:"role" binding:"required"`
	EmployeeNumber string      `json:"employee_number,omitempty" binding:"omitempty,max=20"`
	Department     string      `json:"department,omitempty" binding:"omitempty,max=50"`
}

// Create は新しいユーザーを作成します
func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := validator.ValidatePassword(req.Password, false); err != nil {
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

	// ロールをバリデーション
	if req.Role != domain.RoleAdmin && req.Role != domain.RoleEditor && req.Role != domain.RoleViewer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	user, err := h.userUsecase.CreateUser(c.Request.Context(), req.Email, req.Password, req.Name, req.Role, req.EmployeeNumber, req.Department)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUserRequest はユーザー更新の request body を表します
type UpdateUserRequest struct {
	Name           string      `json:"name" binding:"required,max=50"`
	Email          string      `json:"email" binding:"required,email,max=254"`
	Role           domain.Role `json:"role" binding:"required"`
	EmployeeNumber string      `json:"employee_number,omitempty" binding:"omitempty,max=20"`
	Department     string      `json:"department,omitempty" binding:"omitempty,max=50"`
}

// Update は既存のユーザーを更新します
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if err := validator.ValidateEmployeeNumber(req.EmployeeNumber); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validator.ValidateDepartment(req.Department); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ロールをバリデーション
	if req.Role != domain.RoleAdmin && req.Role != domain.RoleEditor && req.Role != domain.RoleViewer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	user, err := h.userUsecase.Update(c.Request.Context(), uint(id), req.Name, req.Email, req.Role, req.EmployeeNumber, req.Department)
	if err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdatePasswordRequest はパスワード更新の request body を表します
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UpdatePassword はユーザーのパスワードを更新します
func (h *UserHandler) UpdatePassword(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 新しいパスワードのカスタムバリデーション
	if err := validator.ValidatePassword(req.NewPassword, true); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUsecase.UpdatePassword(c.Request.Context(), uint(id), req.OldPassword, req.NewPassword); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

// AdminResetPasswordRequest は管理者によるパスワードリセットの request body を表します
type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// AdminResetPassword は管理者がユーザーのパスワードをリセットします
func (h *UserHandler) AdminResetPassword(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 新しいパスワードのカスタムバリデーション（管理者はより緩い要件）
	if err := validator.ValidatePassword(req.NewPassword, false); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userUsecase.AdminResetPassword(c.Request.Context(), uint(id), req.NewPassword); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset successfully"})
}

// Delete は指定されたユーザーを削除します
func (h *UserHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	if err := h.userUsecase.Delete(c.Request.Context(), uint(id)); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// ToggleActiveRequest はユーザー有効/無効切替の request body を表します
type ToggleActiveRequest struct {
	IsActive bool `json:"is_active"`
}

// ToggleActive はユーザーの有効/無効状態を切り替えます
func (h *UserHandler) ToggleActive(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req ToggleActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// context から現在のユーザー ID を取得
	currentUserIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}
	currentUserID, ok := currentUserIDValue.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format in context"})
		return
	}

	if err := h.userUsecase.ToggleActive(c.Request.Context(), currentUserID, uint(id), req.IsActive); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user status updated successfully"})
}
