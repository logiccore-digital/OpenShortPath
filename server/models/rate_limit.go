package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RateLimit represents a rate limit record in the database
type RateLimit struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Identifier   string    `gorm:"index;size:255;not null" json:"identifier"` // IP address or user_id
	Type         string    `gorm:"index;size:20;not null" json:"type"`        // "ip" or "user"
	RequestCount int       `gorm:"default:0;not null" json:"request_count"`
	WindowStart  time.Time `gorm:"index;not null" json:"window_start"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (RateLimit) TableName() string {
	return "rate_limits"
}

// BeforeCreate hook to generate UUID
func (r *RateLimit) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

