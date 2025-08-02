package routes

import (
	"database/sql"

	"github.com/connect-up/auth-service/handlers"
	"github.com/connect-up/auth-service/utils"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes sets up authentication routes
func SetupAuthRoutes(router *gin.Engine, db *sql.DB) {
	authHandler := handlers.NewAuthHandler(db)

	// Public routes (no authentication required)
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes (authentication required)
	protected := router.Group("/auth")
	protected.Use(utils.AuthMiddleware())
	{
		protected.POST("/logout", authHandler.Logout)
		protected.GET("/profile", authHandler.GetProfile)
	}
} 