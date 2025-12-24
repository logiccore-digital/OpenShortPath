package models

import (
	"time"
)

// Namespace represents a namespace for organizing short URLs
// Namespaces enable URL patterns like domain.com/namespace/slug
type Namespace struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	Name      string    `gorm:"uniqueIndex:idx_domain_name;size:255;not null" json:"name"`
	Domain    string    `gorm:"uniqueIndex:idx_domain_name;size:255;not null" json:"domain"`
	UserID    string    `gorm:"index;size:255;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Namespace) TableName() string {
	return "namespaces"
}

