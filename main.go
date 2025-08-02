package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/connect-up/auth-service/handlers"
	"github.com/connect-up/auth-service/internal/matchmaker"
	"github.com/connect-up/auth-service/models"
	"github.com/connect-up/auth-service/routes"
	"github.com/connect-up/auth-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize JWT
	utils.InitJWT()

	// Initialize database
	if err := models.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	if err := utils.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	// Create Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Initialize matchmaker service
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafkaTopic := getEnv("KAFKA_USER_UPDATED_TOPIC", "user-updated")
	
	matchmakerService := matchmaker.NewService(kafkaBrokers, kafkaTopic)
	defer matchmakerService.Close()

	// Start Kafka consumer in background
	go func() {
		ctx := context.Background()
		matchmakerService.StartConsumer(ctx)
	}()

	// Initialize matchmaker handler
	matchmakerHandler := handlers.NewMatchmakerHandler(matchmakerService)

	// Setup routes
	routes.SetupAuthRoutes(router, models.DB)
	routes.SetupMatchmakerRoutes(router, matchmakerHandler)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "auth-service",
		})
	})

	// Get port from environment or use default
	port := getEnv("PORT", "8080")

	log.Printf("Auth service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
