package matchmaker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/connect-up/auth-service/models"
	"github.com/connect-up/auth-service/utils"
)

type Service struct {
	reader *kafka.Reader
	writer *kafka.Writer
}

// NewService creates a new matchmaker service
func NewService(kafkaBrokers []string, topic string) *Service {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaBrokers,
		Topic:    topic,
		GroupID:  "matchmaker-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBrokers...),
		Topic:    "matches-created",
		Balancer: &kafka.LeastBytes{},
	}

	return &Service{
		reader: reader,
		writer: writer,
	}
}

// StartConsumer starts the Kafka consumer for user-updated events
func (s *Service) StartConsumer(ctx context.Context) {
	log.Println("Starting matchmaker Kafka consumer...")

	for {
		m, err := s.reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		var event models.UserUpdatedEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		log.Printf("Processing user update for user: %s", event.UserID)
		if err := s.ProcessUserUpdate(ctx, event); err != nil {
			log.Printf("Error processing user update: %v", err)
		}
	}
}

// ProcessUserUpdate processes a user update event and finds matches
func (s *Service) ProcessUserUpdate(ctx context.Context, event models.UserUpdatedEvent) error {
	// Store the updated profile
	if err := s.StoreUserProfile(ctx, event.Profile); err != nil {
		return fmt.Errorf("failed to store user profile: %v", err)
	}

	// Find matches for the updated user
	matches, err := s.FindMatches(ctx, event.UserID)
	if err != nil {
		return fmt.Errorf("failed to find matches: %v", err)
	}

	// Store matches
	for _, match := range matches {
		if err := s.StoreMatch(ctx, match); err != nil {
			log.Printf("Failed to store match: %v", err)
			continue
		}
	}

	// Publish match creation events
	if len(matches) > 0 {
		if err := s.PublishMatchesCreated(ctx, matches); err != nil {
			log.Printf("Failed to publish matches created: %v", err)
		}
	}

	return nil
}

// StoreUserProfile stores a user profile in Redis
func (s *Service) StoreUserProfile(ctx context.Context, profile models.UserProfile) error {
	key := fmt.Sprintf("user_profile:%s", profile.UserID)
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	return utils.RedisClient.Set(ctx, key, data, 24*time.Hour).Err()
}

// GetUserProfile retrieves a user profile from Redis
func (s *Service) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	key := fmt.Sprintf("user_profile:%s", userID)
	data, err := utils.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var profile models.UserProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// FindMatches finds potential matches for a user
func (s *Service) FindMatches(ctx context.Context, userID string) ([]models.Match, error) {
	userProfile, err := s.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %v", err)
	}

	// Get all user profiles
	profiles, err := s.GetAllUserProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all profiles: %v", err)
	}

	var matches []models.Match
	for _, profile := range profiles {
		if profile.UserID == userID {
			continue // Skip self
		}

		score := s.CalculateMatchScore(userProfile, &profile)
		if score > 0.3 { // Minimum match threshold
			match := models.Match{
				ID:           uuid.New().String(),
				UserID1:      userID,
				UserID2:      profile.UserID,
				Score:        score,
				CommonTags:   s.FindCommonTags(userProfile.Tags, profile.Tags),
				CommonSkills: s.FindCommonSkills(userProfile.Skills, profile.Skills),
				Status:       "pending",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			matches = append(matches, match)
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Limit to top 10 matches
	if len(matches) > 10 {
		matches = matches[:10]
	}

	return matches, nil
}

// CalculateMatchScore calculates a match score between two users
func (s *Service) CalculateMatchScore(profile1, profile2 *models.UserProfile) float64 {
	var score float64
	var totalWeight float64

	// Tag similarity (weight: 0.3)
	tagScore := s.calculateSimilarity(profile1.Tags, profile2.Tags)
	score += tagScore * 0.3
	totalWeight += 0.3

	// Industry similarity (weight: 0.25)
	industryScore := s.calculateSimilarity(profile1.Industries, profile2.Industries)
	score += industryScore * 0.25
	totalWeight += 0.25

	// Experience compatibility (weight: 0.2)
	expScore := s.calculateExperienceCompatibility(profile1.Experience, profile2.Experience)
	score += expScore * 0.2
	totalWeight += 0.2

	// Skills similarity (weight: 0.15)
	skillsScore := s.calculateSimilarity(profile1.Skills, profile2.Skills)
	score += skillsScore * 0.15
	totalWeight += 0.15

	// Location similarity (weight: 0.1)
	locationScore := s.calculateLocationCompatibility(profile1.Location, profile2.Location)
	score += locationScore * 0.1
	totalWeight += 0.1

	return score / totalWeight
}

// calculateSimilarity calculates Jaccard similarity between two string slices
func (s *Service) calculateSimilarity(slice1, slice2 []string) float64 {
	if len(slice1) == 0 && len(slice2) == 0 {
		return 1.0
	}
	if len(slice1) == 0 || len(slice2) == 0 {
		return 0.0
	}

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, item := range slice1 {
		set1[strings.ToLower(item)] = true
	}
	for _, item := range slice2 {
		set2[strings.ToLower(item)] = true
	}

	intersection := 0
	for item := range set1 {
		if set2[item] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// calculateExperienceCompatibility calculates compatibility based on experience levels
func (s *Service) calculateExperienceCompatibility(exp1, exp2 int) float64 {
	diff := math.Abs(float64(exp1 - exp2))
	if diff <= 2 {
		return 1.0
	} else if diff <= 5 {
		return 0.7
	} else if diff <= 10 {
		return 0.4
	}
	return 0.1
}

// calculateLocationCompatibility calculates location compatibility
func (s *Service) calculateLocationCompatibility(loc1, loc2 string) float64 {
	if loc1 == "" || loc2 == "" {
		return 0.5 // Neutral score for missing location
	}

	loc1Lower := strings.ToLower(strings.TrimSpace(loc1))
	loc2Lower := strings.ToLower(strings.TrimSpace(loc2))

	if loc1Lower == loc2Lower {
		return 1.0
	}

	// Simple city/state matching
	parts1 := strings.Split(loc1Lower, ",")
	parts2 := strings.Split(loc2Lower, ",")

	for _, part1 := range parts1 {
		for _, part2 := range parts2 {
			if strings.TrimSpace(part1) == strings.TrimSpace(part2) {
				return 0.8
			}
		}
	}

	return 0.2
}

// FindCommonTags finds common tags between two users
func (s *Service) FindCommonTags(tags1, tags2 []string) []string {
	set1 := make(map[string]bool)
	for _, tag := range tags1 {
		set1[strings.ToLower(tag)] = true
	}

	var common []string
	for _, tag := range tags2 {
		if set1[strings.ToLower(tag)] {
			common = append(common, tag)
		}
	}

	return common
}

// FindCommonSkills finds common skills between two users
func (s *Service) FindCommonSkills(skills1, skills2 []string) []string {
	set1 := make(map[string]bool)
	for _, skill := range skills1 {
		set1[strings.ToLower(skill)] = true
	}

	var common []string
	for _, skill := range skills2 {
		if set1[strings.ToLower(skill)] {
			common = append(common, skill)
		}
	}

	return common
}

// GetAllUserProfiles retrieves all user profiles from Redis
func (s *Service) GetAllUserProfiles(ctx context.Context) ([]models.UserProfile, error) {
	pattern := "user_profile:*"
	keys, err := utils.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var profiles []models.UserProfile
	for _, key := range keys {
		data, err := utils.RedisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var profile models.UserProfile
		if err := json.Unmarshal([]byte(data), &profile); err != nil {
			continue
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// StoreMatch stores a match in Redis
func (s *Service) StoreMatch(ctx context.Context, match models.Match) error {
	key := fmt.Sprintf("match:%s", match.ID)
	data, err := json.Marshal(match)
	if err != nil {
		return err
	}

	return utils.RedisClient.Set(ctx, key, data, 7*24*time.Hour).Err()
}

// GetMatchesForUser retrieves matches for a specific user
func (s *Service) GetMatchesForUser(ctx context.Context, userID string) ([]models.Match, error) {
	pattern := "match:*"
	keys, err := utils.RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var matches []models.Match
	for _, key := range keys {
		data, err := utils.RedisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var match models.Match
		if err := json.Unmarshal([]byte(data), &match); err != nil {
			continue
		}

		if match.UserID1 == userID || match.UserID2 == userID {
			matches = append(matches, match)
		}
	}

	// Sort by score descending
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	return matches, nil
}

// PublishMatchesCreated publishes match creation events to Kafka
func (s *Service) PublishMatchesCreated(ctx context.Context, matches []models.Match) error {
	for _, match := range matches {
		data, err := json.Marshal(match)
		if err != nil {
			continue
		}

		err = s.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(match.ID),
			Value: data,
		})
		if err != nil {
			log.Printf("Failed to publish match created event: %v", err)
		}
	}

	return nil
}

// Close closes the Kafka connections
func (s *Service) Close() error {
	if s.reader != nil {
		if err := s.reader.Close(); err != nil {
			return err
		}
	}
	if s.writer != nil {
		if err := s.writer.Close(); err != nil {
			return err
		}
	}
	return nil
}
