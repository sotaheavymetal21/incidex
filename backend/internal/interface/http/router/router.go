package router

import (
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware) {
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Example protected route for verification
		api.GET("/protected", jwtMiddleware.Handle(), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			c.JSON(200, gin.H{"message": "You are authorized", "user_id": userID, "role": role})
		})
	}
}
