package router

import (
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware, tagHandler *handler.TagHandler, incidentHandler *handler.IncidentHandler, userHandler *handler.UserHandler, statsHandler *handler.StatsHandler, activityHandler *handler.IncidentActivityHandler, exportHandler *handler.ExportHandler, attachmentHandler *handler.AttachmentHandler, notificationHandler *handler.NotificationHandler, templateHandler *handler.IncidentTemplateHandler) {
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

			// Incident routes
			incidents := protected.Group("/incidents")
			{
				incidents.POST("", incidentHandler.Create)
				incidents.GET("", incidentHandler.GetAll)
				incidents.GET("/:id", incidentHandler.GetByID)
				incidents.PUT("/:id", incidentHandler.Update)
				incidents.DELETE("/:id", incidentHandler.Delete)

				// Incident activity routes
				incidents.POST("/:id/comments", activityHandler.AddComment)
				incidents.GET("/:id/activities", activityHandler.GetActivities)

				// Incident attachment routes
				incidents.POST("/:id/attachments", attachmentHandler.Upload)
				incidents.GET("/:id/attachments", attachmentHandler.GetByIncidentID)
				incidents.GET("/:id/attachments/:attachmentId", attachmentHandler.Download)
				incidents.DELETE("/:id/attachments/:attachmentId", attachmentHandler.Delete)
			}

			// User routes
			protected.GET("/users", userHandler.GetAll)

			// Stats routes
			stats := protected.Group("/stats")
			{
				stats.GET("/dashboard", statsHandler.GetDashboardStats)
				stats.GET("/sla", statsHandler.GetSLAMetrics)
			}

			// Export routes
			export := protected.Group("/export")
			{
				export.GET("/incidents", exportHandler.ExportIncidentsCSV)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/settings", notificationHandler.GetMyNotificationSetting)
				notifications.PUT("/settings", notificationHandler.UpdateMyNotificationSetting)
				notifications.GET("/settings/:id", notificationHandler.GetUserNotificationSetting)
			}

			// Template routes
			templates := protected.Group("/templates")
			{
				templates.POST("", templateHandler.Create)
				templates.GET("", templateHandler.GetAll)
				templates.GET("/:id", templateHandler.GetByID)
				templates.PUT("/:id", templateHandler.Update)
				templates.DELETE("/:id", templateHandler.Delete)
				templates.POST("/create-incident", templateHandler.CreateIncidentFromTemplate)
			}
		}
	}
}
