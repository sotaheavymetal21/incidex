package handler

import (
	"fmt"
	"incidex/internal/domain"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseIDParam はURLパラメータからIDをパースします
func ParseIDParam(c *gin.Context, paramName string) (uint, error) {
	idParam := c.Param(paramName)
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return 0, domain.ErrBadRequest(fmt.Sprintf("invalid %s", paramName))
	}
	return uint(id), nil
}

// GetUserIDFromContext はコンテキストからユーザーIDを取得します
func GetUserIDFromContext(c *gin.Context) (uint, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return 0, domain.ErrUnauthorized("User not authenticated")
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		return 0, domain.ErrInternal("Failed to get user ID", nil)
	}

	return userID, nil
}

// GetUserRoleFromContext はコンテキストからユーザーロールを取得します
func GetUserRoleFromContext(c *gin.Context) (domain.Role, error) {
	roleValue, exists := c.Get("role")
	if !exists {
		return "", domain.ErrUnauthorized("User role not found")
	}

	role, ok := roleValue.(domain.Role)
	if !ok {
		return "", domain.ErrInternal("Invalid user role type", nil)
	}

	return role, nil
}
