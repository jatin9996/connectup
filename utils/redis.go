package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis initializes the Redis connection
func InitRedis() error {
	// Get Redis connection details from environment
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnv("REDIS_DB", "0")

	// Create Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0, // Use default DB
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

// StoreToken stores a token in Redis with expiration
func StoreToken(ctx context.Context, key, token string, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, token, expiration).Err()
}

// GetToken retrieves a token from Redis
func GetToken(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// DeleteToken deletes a token from Redis
func DeleteToken(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// StoreRefreshToken stores a refresh token in Redis
func StoreRefreshToken(ctx context.Context, userID, refreshToken string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return StoreToken(ctx, key, refreshToken, expiration)
}

// GetRefreshToken retrieves a refresh token from Redis
func GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return GetToken(ctx, key)
}

// DeleteRefreshToken deletes a refresh token from Redis
func DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return DeleteToken(ctx, key)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
