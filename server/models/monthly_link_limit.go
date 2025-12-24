package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MonthlyLinkLimit represents a monthly link limit record in the database
type MonthlyLinkLimit struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Identifier  string    `gorm:"index;size:255;not null" json:"identifier"` // IP address or user_id
	Type        string    `gorm:"index;size:20;not null" json:"type"`        // "ip" or "user"
	LinkCount   int       `gorm:"default:0;not null" json:"link_count"`
	MonthStart  time.Time `gorm:"index;not null" json:"month_start"` // First day of the month
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (MonthlyLinkLimit) TableName() string {
	return "monthly_link_limits"
}

// BeforeCreate hook to generate UUID
func (m *MonthlyLinkLimit) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

