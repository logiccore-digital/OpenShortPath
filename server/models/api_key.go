package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// StringArray is a custom type for handling JSON arrays in PostgreSQL
type StringArray []string

// Value implements driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return json.Unmarshal([]byte("[]"), a)
	}
	
	return json.Unmarshal(bytes, a)
}

// APIKey represents an API key in the database
type APIKey struct {
	ID        string     `gorm:"primaryKey;size:36" json:"id"`
	UserID    string     `gorm:"index;size:255;not null" json:"user_id"`
	HashedKey string     `gorm:"size:255;not null" json:"-"` // Never serialize key hash
	Scopes    StringArray `gorm:"type:json" json:"scopes"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (APIKey) TableName() string {
	return "api_keys"
}

