package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/connect-up/auth-service/internal/matchmaker"
	"github.com/connect-up/auth-service/models"
	"github.com/connect-up/auth-service/utils"
	"github.com/gin-gonic/gin"
)

type MatchmakerHandler struct {
	matchmakerService *matchmaker.Service
}

func NewMatchmakerHandler(matchmakerService *matchmaker.Service) *MatchmakerHandler {
	return &MatchmakerHandler{
		matchmakerService: matchmakerService,
	}
}

// CreateUserProfile creates a new user profile for matchmaking
func (h *MatchmakerHandler) CreateUserProfile(c *gin.Context) {
	var req models.MatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile := models.UserProfile{
		UserID:     req.UserID,
		Tags:       req.Tags,
		Industries: req.Industries,
		Experience: req.Experience,
		Interests:  req.Interests,
		Location:   req.Location,
		Bio:        req.Bio,
		Skills:     req.Skills,
	}

	if err := h.matchmakerService.StoreUserProfile(c.Request.Context(), profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user profile"})
		return
	}

	// Trigger match finding
	matches, err := h.matchmakerService.FindMatches(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find matches"})
		return
	}

	// Store matches
	for _, match := range matches {
		if err := h.matchmakerService.StoreMatch(c.Request.Context(), match); err != nil {
			continue
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User profile created successfully",
		"matches_found": len(matches),
	})
}

// GetUserProfile retrieves a user profile
func (h *MatchmakerHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	profile, err := h.matchmakerService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": profile})
}

// GetMatches retrieves matches for a user
func (h *MatchmakerHandler) GetMatches(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Get query parameters for filtering
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	matches, err := h.matchmakerService.GetMatchesForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve matches"})
		return
	}

	// Filter by status if provided
	if status != "" {
		var filteredMatches []models.Match
		for _, match := range matches {
			if match.Status == status {
				filteredMatches = append(filteredMatches, match)
			}
		}
		matches = filteredMatches
	}

	// Apply pagination
	total := len(matches)
	if offset >= total {
		matches = []models.Match{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		matches = matches[offset:end]
	}

	response := models.MatchResponse{
		Matches: matches,
		Total:   total,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateMatchStatus updates the status of a match
func (h *MatchmakerHandler) UpdateMatchStatus(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Match ID is required"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending accepted rejected"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the match from Redis
	key := "match:" + matchID
	data, err := utils.RedisClient.Get(c.Request.Context(), key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	var match models.Match
	if err := json.Unmarshal([]byte(data), &match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse match data"})
		return
	}

	// Update status
	match.Status = req.Status
	match.UpdatedAt = time.Now()

	// Store updated match
	if err := h.matchmakerService.StoreMatch(c.Request.Context(), match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Match status updated successfully",
		"match":   match,
	})
}

// GetMatchDetails retrieves details of a specific match
func (h *MatchmakerHandler) GetMatchDetails(c *gin.Context) {
	matchID := c.Param("match_id")
	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Match ID is required"})
		return
	}

	key := "match:" + matchID
	data, err := utils.RedisClient.Get(c.Request.Context(), key).Result()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
		return
	}

	var match models.Match
	if err := json.Unmarshal([]byte(data), &match); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse match data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"match": match})
}

// SearchMatches searches for matches based on criteria
func (h *MatchmakerHandler) SearchMatches(c *gin.Context) {
	var criteria models.MatchmakingCriteria
	if err := c.ShouldBindJSON(&criteria); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get all profiles
	profiles, err := h.matchmakerService.GetAllUserProfiles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profiles"})
		return
	}

	var matches []models.MatchScore
	userProfile, err := h.matchmakerService.GetUserProfile(c.Request.Context(), criteria.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
		return
	}

	for _, profile := range profiles {
		if profile.UserID == criteria.UserID {
			continue // Skip self
		}

		// Apply filters
		if !h.matchesCriteria(&profile, &criteria) {
			continue
		}

		score := h.matchmakerService.CalculateMatchScore(userProfile, &profile)
		if score > 0.3 { // Minimum threshold
			matches = append(matches, models.MatchScore{
				UserID: profile.UserID,
				Score:  score,
				Reason: h.generateMatchReason(userProfile, &profile),
			})
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Apply pagination
	if criteria.Limit > 0 && criteria.Offset < len(matches) {
		end := criteria.Offset + criteria.Limit
		if end > len(matches) {
			end = len(matches)
		}
		matches = matches[criteria.Offset:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
		"total":   len(matches),
	})
}

// matchesCriteria checks if a profile matches the search criteria
func (h *MatchmakerHandler) matchesCriteria(profile *models.UserProfile, criteria *models.MatchmakingCriteria) bool {
	// Check industries
	if len(criteria.Industries) > 0 {
		found := false
		for _, industry := range criteria.Industries {
			for _, profileIndustry := range profile.Industries {
				if strings.ToLower(industry) == strings.ToLower(profileIndustry) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check experience range
	if criteria.MinExp > 0 && profile.Experience < criteria.MinExp {
		return false
	}
	if criteria.MaxExp > 0 && profile.Experience > criteria.MaxExp {
		return false
	}

	// Check skills
	if len(criteria.Skills) > 0 {
		found := false
		for _, skill := range criteria.Skills {
			for _, profileSkill := range profile.Skills {
				if strings.ToLower(skill) == strings.ToLower(profileSkill) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check location
	if criteria.Location != "" {
		if !strings.Contains(strings.ToLower(profile.Location), strings.ToLower(criteria.Location)) {
			return false
		}
	}

	return true
}

// generateMatchReason generates a reason for the match
func (h *MatchmakerHandler) generateMatchReason(profile1, profile2 *models.UserProfile) string {
	var reasons []string

	// Check common tags
	commonTags := h.matchmakerService.FindCommonTags(profile1.Tags, profile2.Tags)
	if len(commonTags) > 0 {
		reasons = append(reasons, fmt.Sprintf("Common interests: %s", strings.Join(commonTags, ", ")))
	}

	// Check common skills
	commonSkills := h.matchmakerService.FindCommonSkills(profile1.Skills, profile2.Skills)
	if len(commonSkills) > 0 {
		reasons = append(reasons, fmt.Sprintf("Common skills: %s", strings.Join(commonSkills, ", ")))
	}

	// Check experience compatibility
	expDiff := abs(profile1.Experience - profile2.Experience)
	if expDiff <= 2 {
		reasons = append(reasons, "Similar experience level")
	}

	// Check location
	if profile1.Location != "" && profile2.Location != "" {
		if strings.ToLower(profile1.Location) == strings.ToLower(profile2.Location) {
			reasons = append(reasons, "Same location")
		}
	}

	if len(reasons) == 0 {
		return "Good overall compatibility"
	}

	return strings.Join(reasons, "; ")
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
} 