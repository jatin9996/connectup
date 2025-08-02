package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/connect-up/auth-service/models"
	"github.com/connect-up/auth-service/utils"
)

func main() {
	// Initialize Kafka producer
	brokers := []string{"localhost:9092"}
	topic := "user-updated"
	producer := utils.NewKafkaProducer(brokers, topic)
	defer producer.Close()

	// Create sample user profiles
	users := []struct {
		ID     string
		Profile models.UserProfile
	}{
		{
			ID: "user1",
			Profile: models.UserProfile{
				UserID:     "user1",
				Tags:       []string{"golang", "backend", "microservices", "docker"},
				Industries: []string{"technology", "software", "fintech"},
				Experience: 5,
				Interests:  []string{"open source", "cloud computing", "distributed systems"},
				Location:   "San Francisco, CA",
				Bio:        "Backend developer with 5 years of experience in Go and microservices",
				Skills:     []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			ID: "user2",
			Profile: models.UserProfile{
				UserID:     "user2",
				Tags:       []string{"golang", "backend", "api", "testing"},
				Industries: []string{"technology", "software", "ecommerce"},
				Experience: 3,
				Interests:  []string{"testing", "api design", "performance"},
				Location:   "San Francisco, CA",
				Bio:        "Backend developer focused on API design and testing",
				Skills:     []string{"Go", "PostgreSQL", "Redis", "Testing", "API Design"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			ID: "user3",
			Profile: models.UserProfile{
				UserID:     "user3",
				Tags:       []string{"frontend", "react", "javascript", "ui"},
				Industries: []string{"technology", "software", "healthcare"},
				Experience: 4,
				Interests:  []string{"user experience", "design systems", "accessibility"},
				Location:   "New York, NY",
				Bio:        "Frontend developer with expertise in React and UI design",
				Skills:     []string{"React", "JavaScript", "TypeScript", "CSS", "UI/UX"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		{
			ID: "user4",
			Profile: models.UserProfile{
				UserID:     "user4",
				Tags:       []string{"golang", "backend", "microservices", "kubernetes"},
				Industries: []string{"technology", "software", "fintech"},
				Experience: 7,
				Interests:  []string{"distributed systems", "cloud native", "observability"},
				Location:   "San Francisco, CA",
				Bio:        "Senior backend engineer with expertise in distributed systems",
				Skills:     []string{"Go", "Kubernetes", "Docker", "Microservices", "Monitoring"},
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
	}

	ctx := context.Background()

	// Publish user updated events
	fmt.Println("Publishing user updated events...")
	for _, user := range users {
		if err := producer.PublishUserUpdated(ctx, user.ID, user.Profile); err != nil {
			log.Printf("Failed to publish event for user %s: %v", user.ID, err)
		} else {
			fmt.Printf("Published event for user: %s\n", user.ID)
		}
		time.Sleep(1 * time.Second) // Small delay between events
	}

	fmt.Println("\nUser events published successfully!")
	fmt.Println("The matchmaker service should now process these events and create matches.")
	fmt.Println("\nYou can test the REST endpoints:")
	fmt.Println("1. Get matches for user1: GET http://localhost:8080/api/v1/matchmaker/matches/user1")
	fmt.Println("2. Get user profile: GET http://localhost:8080/api/v1/matchmaker/profiles/user1")
	fmt.Println("3. Search matches: POST http://localhost:8080/api/v1/matchmaker/search")
	fmt.Println("   Body: {\"user_id\": \"user1\", \"limit\": 10, \"offset\": 0}")
} 