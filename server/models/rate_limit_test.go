package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRateLimit_BeforeCreate_GeneratesUUID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	db.AutoMigrate(&RateLimit{})

	rateLimit := &RateLimit{
		Identifier:   "test-identifier",
		Type:         "ip",
		RequestCount: 1,
		WindowStart:  time.Now().UTC().Truncate(time.Hour),
	}

	err = db.Create(rateLimit).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, rateLimit.ID)
	
	// Verify it's a valid UUID
	_, err = uuid.Parse(rateLimit.ID)
	assert.NoError(t, err)
}

func TestRateLimit_BeforeCreate_PreservesExistingUUID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	db.AutoMigrate(&RateLimit{})

	existingID := uuid.New().String()
	rateLimit := &RateLimit{
		ID:           existingID,
		Identifier:   "test-identifier",
		Type:         "user",
		RequestCount: 1,
		WindowStart:  time.Now().UTC().Truncate(time.Hour),
	}

	err = db.Create(rateLimit).Error
	assert.NoError(t, err)
	assert.Equal(t, existingID, rateLimit.ID)
}

func TestRateLimit_TableName(t *testing.T) {
	var r RateLimit
	assert.Equal(t, "rate_limits", r.TableName())
}

