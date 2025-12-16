package router

import (
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware, tagHandler *handler.TagHandler) {
	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(jwtMiddleware.Handle())
		{
			protected.GET("/protected", func(c *gin.Context) {
				userID, _ := c.Get("userID")
				role, _ := c.Get("role")
				c.JSON(200, gin.H{
					"message": "You are logged in",
					"userID":  userID,
					"role":    role,
				})
			})

			// Tag routes
			tags := protected.Group("/tags")
			{
				tags.POST("", tagHandler.Create)
				tags.GET("", tagHandler.GetAll)
				tags.PUT("/:id", tagHandler.Update)
				tags.DELETE("/:id", tagHandler.Delete)
			}
		}
	}
}
