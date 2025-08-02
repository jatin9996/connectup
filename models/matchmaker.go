package models

import (
	"time"
type UserProfile struct {
	UserID     string    `json:"user_id" db:"user_id"`
	Tags       []string  `json:"tags" db:"tags"`
	Industries []string  `json:"industries" db:"industries"`
	Experience int       `json:"experience" db:"experience"` // years of experience
	Interests  []string  `json:"interests" db:"interests"`
	Location   string    `json:"location" db:"location"`
	Bio        string    `json:"bio" db:"bio"`
	Skills     []string  `json:"skills" db:"skills"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Match represents a match between two users
type Match struct {
	ID           string    `json:"id" db:"id"`
	UserID1      string    `json:"user_id_1" db:"user_id_1"`
	UserID2      string    `json:"user_id_2" db:"user_id_2"`
	Score        float64   `json:"score" db:"score"`
	CommonTags   []string  `json:"common_tags" db:"common_tags"`
	CommonSkills []string  `json:"common_skills" db:"common_skills"`
	Status       string    `json:"status" db:"status"` // pending, accepted, rejected
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// MatchRequest represents the request to create a user profile
type MatchRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	Tags       []string `json:"tags"`
	Industries []string `json:"industries"`
	Experience int      `json:"experience"`
	Interests  []string `json:"interests"`
	Location   string   `json:"location"`
	Bio        string   `json:"bio"`
	Skills     []string `json:"skills"`
}

// MatchResponse represents the response for match endpoints
type MatchResponse struct {
	Matches []Match `json:"matches"`
	Total   int     `json:"total"`
}

// UserUpdatedEvent represents the Kafka event for user updates
type UserUpdatedEvent struct {
	UserID    string      `json:"user_id"`
	Profile   UserProfile `json:"profile"`
	Timestamp time.Time   `json:"timestamp"`
}

// MatchScore represents a match score calculation
type MatchScore struct {
	UserID string  `json:"user_id"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

// MatchmakingCriteria represents the criteria for finding matches
type MatchmakingCriteria struct {
	UserID     string   `json:"user_id"`
	Tags       []string `json:"tags"`
	Industries []string `json:"industries"`
	MinExp     int      `json:"min_exp"`
	MaxExp     int      `json:"max_exp"`
	Skills     []string `json:"skills"`
	Location   string   `json:"location"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}
