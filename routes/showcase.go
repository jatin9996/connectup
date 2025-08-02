package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/connect-up/auth-service/handlers"
	"github.com/connect-up/auth-service/utils"
)

// SetupShowcaseRoutes sets up the showcase service routes
func SetupShowcaseRoutes(router *gin.Engine, showcaseHandler *handlers.ShowcaseHandler) {
	// Showcase API group with authentication middleware
	showcase := router.Group("/api/v1/showcase")
	showcase.Use(utils.AuthMiddleware())
	{
		// Company management (admin/investor only)
		showcase.POST("/companies", showcaseHandler.CreateCompany)
		showcase.GET("/companies/:id", showcaseHandler.GetCompany)
		showcase.PUT("/companies/:id", showcaseHandler.UpdateCompany)
		showcase.GET("/companies", showcaseHandler.SearchCompanies)

		// Investment management (investor only)
		showcase.POST("/investments", showcaseHandler.CreateInvestment)
		showcase.GET("/companies/:company_id/investments", showcaseHandler.GetInvestments)
		showcase.GET("/investments/my", showcaseHandler.GetUserInvestments)

		// Analytics tracking
		showcase.POST("/analytics/events", showcaseHandler.TrackEvent)
	}

	// Public showcase routes (no authentication required)
	publicShowcase := router.Group("/api/v1/showcase/public")
	{
		// Public company profiles
		publicShowcase.GET("/companies", showcaseHandler.SearchCompanies)
		publicShowcase.GET("/companies/:id", showcaseHandler.GetCompany)
	}
}
