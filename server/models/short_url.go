package models

import (
	"time"
)

// ShortURL represents a shortened URL entry in the database
type ShortURL struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	Domain      string    `gorm:"uniqueIndex:idx_domain_slug;size:255" json:"domain"`
	Slug        string    `gorm:"uniqueIndex:idx_domain_slug;size:255" json:"slug"`
	URL         string    `gorm:"not null;size:2048" json:"url"`
	UserID      string    `gorm:"size:255" json:"user_id"`
	NamespaceID *string   `gorm:"index;size:36" json:"namespace_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (ShortURL) TableName() string {
	return "short_urls"
}
