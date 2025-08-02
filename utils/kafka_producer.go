package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/connect-up/auth-service/models"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer represents a Kafka producer
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: writer,
	}
}

// PublishUserUpdated publishes a user updated event
func (kp *KafkaProducer) PublishUserUpdated(ctx context.Context, userID string, profile models.UserProfile) error {
	event := models.UserUpdatedEvent{
		UserID:    userID,
		Profile:   profile,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	err = kp.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(userID),
		Value: data,
	})
	if err != nil {
		return fmt.Errorf("failed to publish event: %v", err)
	}

	log.Printf("Published user updated event for user: %s", userID)
	return nil
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

// CreateSampleUserProfile creates a sample user profile for testing
func CreateSampleUserProfile(userID string) models.UserProfile {
	return models.UserProfile{
		UserID:     userID,
		Tags:       []string{"golang", "backend", "microservices", "docker"},
		Industries: []string{"technology", "software", "fintech"},
		Experience: 5,
		Interests:  []string{"open source", "cloud computing", "distributed systems"},
		Location:   "San Francisco, CA",
		Bio:        "Backend developer with 5 years of experience in Go and microservices",
		Skills:     []string{"Go", "PostgreSQL", "Redis", "Docker", "Kubernetes"},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
} 