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
	"github.com/segmentio/kafka-go"
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

	// Create showcase tables
	if err := models.CreateShowcaseTables(); err != nil {
		log.Fatalf("Failed to create showcase tables: %v", err)
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

	// Initialize Kafka
	kafkaBrokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	kafkaUserTopic := getEnv("KAFKA_USER_UPDATED_TOPIC", "user-updated")
	kafkaChatTopic := getEnv("KAFKA_CHAT_TOPIC", "chat-messages")
	kafkaAnalyticsTopic := getEnv("KAFKA_ANALYTICS_TOPIC", "analytics_events")

	// Create Kafka writer for analytics
	kafkaWriter := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers...),
		Topic:    kafkaAnalyticsTopic,
		Balancer: &kafka.LeastBytes{},
	}

	// Create Kafka reader for chat messages
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaBrokers,
		Topic:    kafkaChatTopic,
		GroupID:  "auth-service-chat-consumer",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	// Initialize matchmaker service
	matchmakerService := matchmaker.NewService(kafkaBrokers, kafkaUserTopic)
	defer matchmakerService.Close()

	// Start Kafka consumer in background
	go func() {
		ctx := context.Background()
		matchmakerService.StartConsumer(ctx)
	}()

	// Initialize handlers
	matchmakerHandler := handlers.NewMatchmakerHandler(matchmakerService)
	showcaseHandler := handlers.NewShowcaseHandler(models.DB, kafkaWriter, utils.RedisClient)
	websocketHandler := handlers.NewWebSocketHandler(kafkaWriter, kafkaReader, models.DB)

	// Setup routes
	routes.SetupAuthRoutes(router, models.DB)
	routes.SetupMatchmakerRoutes(router, matchmakerHandler)
	routes.SetupShowcaseRoutes(router, showcaseHandler)

	// WebSocket routes
	router.GET("/ws", utils.AuthMiddleware(), websocketHandler.HandleWebSocket)
	router.GET("/api/v1/websocket/online-users", utils.AuthMiddleware(), websocketHandler.GetOnlineUsers)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "auth-service",
			"features": []string{
				"authentication",
				"matchmaking",
				"showcase",
				"websocket-messaging",
				"kafka-integration",
				"redis-caching",
			},
		})
	})

	// Get port from environment or use default
	port := getEnv("PORT", "8080")

	log.Printf("Auth service starting on port %s", port)
	log.Printf("Features enabled: Authentication, Matchmaking, Showcase, WebSocket Messaging, Kafka Integration, Redis Caching")

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
