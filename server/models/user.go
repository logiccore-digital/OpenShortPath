package models

import (
	"time"
)

// User represents a user in the database
// Username and HashedPassword are optional to support external authentication providers
type User struct {
	UserID        string    `gorm:"primaryKey;size:255" json:"user_id"`
	Username      *string   `gorm:"size:255;uniqueIndex" json:"username,omitempty"`
	HashedPassword *string  `gorm:"size:255" json:"-"` // Never serialize password hash
	Active        bool      `gorm:"default:true" json:"active"`
	Plan          string    `gorm:"default:'hobbyist'" json:"plan"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

