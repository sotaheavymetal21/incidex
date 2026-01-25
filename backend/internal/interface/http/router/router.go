package router

import (
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware, tagHandler *handler.TagHandler, incidentHandler *handler.IncidentHandler, userHandler *handler.UserHandler, statsHandler *handler.StatsHandler, activityHandler *handler.IncidentActivityHandler, exportHandler *handler.ExportHandler, attachmentHandler *handler.AttachmentHandler, notificationHandler *handler.NotificationHandler, postMortemHandler *handler.PostMortemHandler, actionItemHandler *handler.ActionItemHandler, auditLogHandler *handler.AuditLogHandler, reportHandler *handler.ReportHandler, healthHandler *handler.HealthHandler, passwordResetHandler *handler.PasswordResetHandler, loginRateLimiter *middleware.RateLimitMiddleware, apiRateLimiter *middleware.RateLimitMiddleware) {
	api := r.Group("/api")
	{
		// Health check routes (no auth required)
		api.GET("/health", healthHandler.Liveness)
		api.GET("/health/ready", healthHandler.Readiness)

		// Auth routes
		auth := api.Group("/auth")
		{
			// Apply rate limiting to authentication endpoints
			auth.POST("/register", loginRateLimiter.Handle(), authHandler.Register)
			auth.POST("/login", loginRateLimiter.Handle(), authHandler.Login)
			auth.POST("/refresh", loginRateLimiter.Handle(), authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/forgot-password", loginRateLimiter.Handle(), passwordResetHandler.RequestPasswordReset)
			auth.POST("/reset-password", loginRateLimiter.Handle(), passwordResetHandler.ResetPassword)
			auth.GET("/validate-reset-token", passwordResetHandler.ValidateToken)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(jwtMiddleware.Handle())
		protected.Use(apiRateLimiter.Handle()) // Apply general API rate limiting
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
				tags.POST("", middleware.RequireEditorOrAdmin(), tagHandler.Create)
				tags.GET("", tagHandler.GetAll)
				tags.PUT("/:id", middleware.RequireEditorOrAdmin(), tagHandler.Update)
				tags.DELETE("/:id", middleware.RequireEditorOrAdmin(), tagHandler.Delete)
			}

			// Incident routes
			incidents := protected.Group("/incidents")
			{
				incidents.POST("", middleware.RequireEditorOrAdmin(), incidentHandler.Create)
				incidents.GET("", incidentHandler.GetAll)
				incidents.GET("/:id", incidentHandler.GetByID)
				incidents.PUT("/:id", middleware.RequireEditorOrAdmin(), incidentHandler.Update)
				incidents.DELETE("/:id", middleware.RequireEditorOrAdmin(), incidentHandler.Delete)
				incidents.POST("/:id/assign", middleware.RequireEditorOrAdmin(), incidentHandler.AssignIncident)

				// Incident activity routes
				incidents.POST("/:id/comments", middleware.RequireEditorOrAdmin(), activityHandler.AddComment)
				incidents.POST("/:id/timeline", middleware.RequireEditorOrAdmin(), activityHandler.AddTimelineEvent)
				incidents.GET("/:id/activities", activityHandler.GetActivities)

				// Incident attachment routes
				incidents.POST("/:id/attachments", middleware.RequireEditorOrAdmin(), attachmentHandler.Upload)
				incidents.GET("/:id/attachments", attachmentHandler.GetByIncidentID)
				incidents.GET("/:id/attachments/:attachmentId/download", attachmentHandler.Download)
				incidents.DELETE("/:id/attachments/:attachmentId", middleware.RequireEditorOrAdmin(), attachmentHandler.Delete)

				// Post-mortem routes under incidents
				incidents.GET("/:id/postmortem", postMortemHandler.GetByIncidentID)
			}

			// User routes (admin only)
			users := protected.Group("/users")
			users.Use(middleware.RequireAdmin())
			{
				users.POST("", userHandler.Create)
				users.GET("", userHandler.GetAll)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", userHandler.Update)
				users.PATCH("/:id/status", userHandler.ToggleActive)
				users.PUT("/:id/password", userHandler.UpdatePassword)
				users.POST("/:id/admin-reset-password", userHandler.AdminResetPassword)
				users.DELETE("/:id", userHandler.Delete)
			}

			// Stats routes
			stats := protected.Group("/stats")
			{
				stats.GET("/dashboard", statsHandler.GetDashboardStats)
				stats.GET("/sla", statsHandler.GetSLAMetrics)
			stats.GET("/tags", statsHandler.GetTagStats)
			}

			// Export routes
			export := protected.Group("/export")
			{
				export.GET("/incidents", exportHandler.ExportIncidentsCSV)
				export.GET("/incidents/:id/pdf", exportHandler.ExportIncidentPDF)
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/settings", notificationHandler.GetMyNotificationSetting)
				notifications.PUT("/settings", notificationHandler.UpdateMyNotificationSetting)
				notifications.GET("/settings/:id", notificationHandler.GetUserNotificationSetting)
			}

			// Post-mortem routes
			postMortems := protected.Group("/post-mortems")
			{
				postMortems.POST("", middleware.RequireEditorOrAdmin(), postMortemHandler.Create)
				postMortems.GET("", postMortemHandler.GetAll)
				postMortems.GET("/:id", postMortemHandler.GetByID)
				postMortems.PUT("/:id", middleware.RequireEditorOrAdmin(), postMortemHandler.Update)
				postMortems.DELETE("/:id", middleware.RequireEditorOrAdmin(), postMortemHandler.Delete)
				postMortems.POST("/:id/publish", middleware.RequireEditorOrAdmin(), postMortemHandler.Publish)
				postMortems.POST("/:id/unpublish", middleware.RequireEditorOrAdmin(), postMortemHandler.Unpublish)
				postMortems.GET("/:id/action-items", actionItemHandler.GetByPostMortemID)
			}

			// Action item routes
			actionItems := protected.Group("/action-items")
			{
				actionItems.POST("", middleware.RequireEditorOrAdmin(), actionItemHandler.Create)
				actionItems.GET("", actionItemHandler.GetAll)
				actionItems.GET("/:id", actionItemHandler.GetByID)
				actionItems.PUT("/:id", middleware.RequireEditorOrAdmin(), actionItemHandler.Update)
				actionItems.DELETE("/:id", middleware.RequireEditorOrAdmin(), actionItemHandler.Delete)
			}

		// Audit log routes (admin only)
		auditLogs := protected.Group("/audit-logs")
		auditLogs.Use(middleware.RequireAdmin())
		{
			auditLogs.GET("", auditLogHandler.GetAll)
			auditLogs.GET("/:id", auditLogHandler.GetByID)
		}

		// Report routes
		reports := protected.Group("/reports")
		{
			reports.GET("/monthly", reportHandler.GetMonthlyReport)
			reports.GET("/custom", reportHandler.GetCustomReport)
		}
	}
	}
}
