package router

import (
	"incidex/internal/interface/http/handler"
	"incidex/internal/interface/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, authHandler *handler.AuthHandler, jwtMiddleware *middleware.JWTMiddleware, tagHandler *handler.TagHandler, incidentHandler *handler.IncidentHandler, userHandler *handler.UserHandler, statsHandler *handler.StatsHandler, activityHandler *handler.IncidentActivityHandler, exportHandler *handler.ExportHandler, attachmentHandler *handler.AttachmentHandler, notificationHandler *handler.NotificationHandler, templateHandler *handler.IncidentTemplateHandler, postMortemHandler *handler.PostMortemHandler, actionItemHandler *handler.ActionItemHandler, auditLogHandler *handler.AuditLogHandler, reportHandler *handler.ReportHandler) {
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
				incidents.POST("/:id/summarize", middleware.RequireEditorOrAdmin(), incidentHandler.RegenerateSummary)
				incidents.POST("/:id/assign", middleware.RequireEditorOrAdmin(), incidentHandler.AssignIncident)

				// Incident activity routes
				incidents.POST("/:id/comments", middleware.RequireEditorOrAdmin(), activityHandler.AddComment)
				incidents.POST("/:id/timeline", middleware.RequireEditorOrAdmin(), activityHandler.AddTimelineEvent)
				incidents.GET("/:id/activities", activityHandler.GetActivities)

				// Incident attachment routes
				incidents.POST("/:id/attachments", middleware.RequireEditorOrAdmin(), attachmentHandler.Upload)
				incidents.GET("/:id/attachments", attachmentHandler.GetByIncidentID)
				incidents.GET("/:id/attachments/:attachmentId", attachmentHandler.Download)
				incidents.DELETE("/:id/attachments/:attachmentId", middleware.RequireEditorOrAdmin(), attachmentHandler.Delete)

				// Post-mortem routes under incidents
				incidents.GET("/:id/postmortem", postMortemHandler.GetByIncidentID)
				incidents.POST("/:id/postmortem/ai-suggestion", middleware.RequireEditorOrAdmin(), postMortemHandler.GenerateAISuggestion)
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
				templates.POST("", middleware.RequireEditorOrAdmin(), templateHandler.Create)
				templates.GET("", templateHandler.GetAll)
				templates.GET("/:id", templateHandler.GetByID)
				templates.PUT("/:id", middleware.RequireEditorOrAdmin(), templateHandler.Update)
				templates.DELETE("/:id", middleware.RequireEditorOrAdmin(), templateHandler.Delete)
				templates.POST("/create-incident", middleware.RequireEditorOrAdmin(), templateHandler.CreateIncidentFromTemplate)
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
