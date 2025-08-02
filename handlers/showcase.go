package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	"github.com/connect-up/auth-service/models"
	"github.com/connect-up/auth-service/utils"
)

// ShowcaseHandler handles showcase-related requests
type ShowcaseHandler struct {
	db          *sql.DB
	kafkaWriter *kafka.Writer
	redisClient *utils.RedisClient
}

// NewShowcaseHandler creates a new showcase handler
func NewShowcaseHandler(db *sql.DB, kafkaWriter *kafka.Writer, redisClient *utils.RedisClient) *ShowcaseHandler {
	return &ShowcaseHandler{
		db:          db,
		kafkaWriter: kafkaWriter,
		redisClient: redisClient,
	}
}

// CreateCompany creates a new company profile (admin/investor only)
func (h *ShowcaseHandler) CreateCompany(c *gin.Context) {
	// Check if user is admin or investor
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check user role (you might want to add a role field to your user model)
	// For now, we'll assume all authenticated users can create companies
	// In production, you should check for admin/investor role

	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set the creator
	company.CreatedBy = userID.(string)
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()

	// Create the company
	if err := models.CreateCompany(&company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}

	// Publish to Kafka for analytics
	h.publishAnalyticsEvent(userID.(string), "company_created", map[string]interface{}{
		"company_id":   company.ID,
		"company_name": company.Name,
	})

	// Cache the company profile
	h.cacheCompanyProfile(&company)

	c.JSON(http.StatusCreated, company)
}

// GetCompany retrieves a company profile
func (h *ShowcaseHandler) GetCompany(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	// Try to get from cache first
	cachedCompany, err := h.getCachedCompanyProfile(companyID)
	if err == nil && cachedCompany != nil {
		c.JSON(http.StatusOK, cachedCompany)
		return
	}

	// Get from database
	company, err := models.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company"})
		return
	}

	// Cache the company profile
	h.cacheCompanyProfile(company)

	// Track analytics
	if userID, exists := c.Get("user_id"); exists {
		h.publishAnalyticsEvent(userID.(string), "company_viewed", map[string]interface{}{
			"company_id": company.ID,
		})
	}

	c.JSON(http.StatusOK, company)
}

// UpdateCompany updates a company profile (admin/creator only)
func (h *ShowcaseHandler) UpdateCompany(c *gin.Context) {
	companyID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get existing company to check permissions
	existingCompany, err := models.GetCompanyByID(companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve company"})
		return
	}

	// Check if user is the creator or admin
	if existingCompany.CreatedBy != userID.(string) {
		// In production, check for admin role here
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this company"})
		return
	}

	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	company.ID = companyID
	company.UpdatedAt = time.Now()

	if err := models.UpdateCompany(&company); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	// Invalidate cache
	h.invalidateCompanyCache(companyID)

	// Publish to Kafka
	h.publishAnalyticsEvent(userID.(string), "company_updated", map[string]interface{}{
		"company_id": company.ID,
	})

	c.JSON(http.StatusOK, company)
}

// SearchCompanies searches for companies with filters
func (h *ShowcaseHandler) SearchCompanies(c *gin.Context) {
	query := c.Query("q")
	industry := c.Query("industry")
	fundingStage := c.Query("funding_stage")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	companies, err := models.SearchCompanies(query, industry, fundingStage, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search companies"})
		return
	}

	// Track search analytics
	if userID, exists := c.Get("user_id"); exists {
		h.publishAnalyticsEvent(userID.(string), "company_search", map[string]interface{}{
			"query":         query,
			"industry":      industry,
			"funding_stage": fundingStage,
			"results_count": len(companies),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": companies,
		"total":     len(companies),
		"limit":     limit,
		"offset":    offset,
	})
}

// CreateInvestment creates a new investment record (investor only)
func (h *ShowcaseHandler) CreateInvestment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var investment models.Investment
	if err := c.ShouldBindJSON(&investment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set investor and timestamps
	investment.InvestorID = userID.(string)
	investment.CreatedAt = time.Now()
	investment.UpdatedAt = time.Now()

	// Create investment in database
	if err := h.createInvestment(&investment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create investment"})
		return
	}

	// Publish to Kafka
	h.publishAnalyticsEvent(userID.(string), "investment_created", map[string]interface{}{
		"investment_id": investment.ID,
		"company_id":    investment.CompanyID,
		"amount":        investment.Amount,
		"currency":      investment.Currency,
	})

	c.JSON(http.StatusCreated, investment)
}

// GetInvestments retrieves investments for a company
func (h *ShowcaseHandler) GetInvestments(c *gin.Context) {
	companyID := c.Param("company_id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID is required"})
		return
	}

	investments, err := h.getInvestmentsByCompany(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve investments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"investments": investments})
}

// GetUserInvestments retrieves investments made by a user
func (h *ShowcaseHandler) GetUserInvestments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	investments, err := h.getInvestmentsByUser(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve investments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"investments": investments})
}

// Analytics tracking
func (h *ShowcaseHandler) TrackEvent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var eventData map[string]interface{}
	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event data"})
		return
	}

	eventType, exists := eventData["event_type"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event type is required"})
		return
	}

	// Remove event_type from data as it's stored separately
	delete(eventData, "event_type")

	// Publish to Kafka
	h.publishAnalyticsEvent(userID.(string), eventType, eventData)

	c.JSON(http.StatusOK, gin.H{"message": "Event tracked successfully"})
}

// Helper methods

func (h *ShowcaseHandler) createInvestment(investment *models.Investment) error {
	query := `
		INSERT INTO investments (company_id, investor_id, amount, currency, investment_type, round, date, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	return h.db.QueryRow(query,
		investment.CompanyID, investment.InvestorID, investment.Amount, investment.Currency,
		investment.InvestmentType, investment.Round, investment.Date, investment.Status, investment.Notes,
	).Scan(&investment.ID, &investment.CreatedAt, &investment.UpdatedAt)
}

func (h *ShowcaseHandler) getInvestmentsByCompany(companyID string) ([]models.Investment, error) {
	query := `
		SELECT id, company_id, investor_id, amount, currency, investment_type, round, date, status, notes, created_at, updated_at
		FROM investments
		WHERE company_id = $1
		ORDER BY date DESC
	`

	rows, err := h.db.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []models.Investment
	for rows.Next() {
		var investment models.Investment
		err := rows.Scan(
			&investment.ID, &investment.CompanyID, &investment.InvestorID, &investment.Amount,
			&investment.Currency, &investment.InvestmentType, &investment.Round, &investment.Date,
			&investment.Status, &investment.Notes, &investment.CreatedAt, &investment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		investments = append(investments, investment)
	}

	return investments, nil
}

func (h *ShowcaseHandler) getInvestmentsByUser(userID string) ([]models.Investment, error) {
	query := `
		SELECT id, company_id, investor_id, amount, currency, investment_type, round, date, status, notes, created_at, updated_at
		FROM investments
		WHERE investor_id = $1
		ORDER BY date DESC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []models.Investment
	for rows.Next() {
		var investment models.Investment
		err := rows.Scan(
			&investment.ID, &investment.CompanyID, &investment.InvestorID, &investment.Amount,
			&investment.Currency, &investment.InvestmentType, &investment.Round, &investment.Date,
			&investment.Status, &investment.Notes, &investment.CreatedAt, &investment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		investments = append(investments, investment)
	}

	return investments, nil
}

func (h *ShowcaseHandler) publishAnalyticsEvent(userID, eventType string, eventData map[string]interface{}) {
	if h.kafkaWriter == nil {
		return
	}

	event := map[string]interface{}{
		"user_id":    userID,
		"event_type": eventType,
		"event_data": eventData,
		"timestamp":  time.Now().Unix(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: "analytics_events",
		Key:   []byte(userID),
		Value: eventJSON,
	})
}

func (h *ShowcaseHandler) cacheCompanyProfile(company *models.Company) {
	if h.redisClient == nil {
		return
	}

	companyJSON, err := json.Marshal(company)
	if err != nil {
		return
	}

	// Cache for 1 hour
	h.redisClient.Set(fmt.Sprintf("company:%s", company.ID), string(companyJSON), time.Hour)
}

func (h *ShowcaseHandler) getCachedCompanyProfile(companyID string) (*models.Company, error) {
	if h.redisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	companyJSON, err := h.redisClient.Get(fmt.Sprintf("company:%s", companyID))
	if err != nil {
		return nil, err
	}

	var company models.Company
	if err := json.Unmarshal([]byte(companyJSON), &company); err != nil {
		return nil, err
	}

	return &company, nil
}

func (h *ShowcaseHandler) invalidateCompanyCache(companyID string) {
	if h.redisClient == nil {
		return
	}

	h.redisClient.Del(fmt.Sprintf("company:%s", companyID))
}
