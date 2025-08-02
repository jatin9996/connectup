package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/connect-up/auth-service/handlers"
)

// SetupMatchmakerRoutes sets up the matchmaker routes
func SetupMatchmakerRoutes(router *gin.Engine, matchmakerHandler *handlers.MatchmakerHandler) {
	// Matchmaker API group
	matchmaker := router.Group("/api/v1/matchmaker")
	{
		// User profile management
		matchmaker.POST("/profiles", matchmakerHandler.CreateUserProfile)
		matchmaker.GET("/profiles/:user_id", matchmakerHandler.GetUserProfile)

		// Match management
		matchmaker.GET("/matches/:user_id", matchmakerHandler.GetMatches)
		matchmaker.GET("/matches/details/:match_id", matchmakerHandler.GetMatchDetails)
		matchmaker.PUT("/matches/:match_id/status", matchmakerHandler.UpdateMatchStatus)

		// Search and discovery
		matchmaker.POST("/search", matchmakerHandler.SearchMatches)
	}
}
