package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, key)
	assert.True(t, strings.HasPrefix(key, APIKeyPrefix), "API key should start with %s", APIKeyPrefix)
}

func TestGenerateAPIKey_Format(t *testing.T) {
	key, err := GenerateAPIKey()
	assert.NoError(t, err)
	
	// Should have prefix
	assert.True(t, strings.HasPrefix(key, APIKeyPrefix))
	
	// Should have content after prefix
	keyWithoutPrefix := strings.TrimPrefix(key, APIKeyPrefix)
	assert.NotEmpty(t, keyWithoutPrefix)
	assert.Greater(t, len(keyWithoutPrefix), 20, "Key should have sufficient length after prefix")
}

func TestGenerateAPIKey_UniqueKeys(t *testing.T) {
	key1, err1 := GenerateAPIKey()
	assert.NoError(t, err1)
	
	key2, err2 := GenerateAPIKey()
	assert.NoError(t, err2)
	
	// Keys should be different (very high probability)
	assert.NotEqual(t, key1, key2)
}

func TestGenerateAPIKey_MultipleGenerations(t *testing.T) {
	keys := make(map[string]bool)
	
	// Generate 100 keys and ensure they're all unique
	for i := 0; i < 100; i++ {
		key, err := GenerateAPIKey()
		assert.NoError(t, err)
		assert.False(t, keys[key], "Duplicate key generated at iteration %d", i)
		keys[key] = true
	}
}

func TestGenerateAPIKey_CanBeHashed(t *testing.T) {
	key, err := GenerateAPIKey()
	assert.NoError(t, err)
	
	// Should be able to hash the key (same as password hashing)
	hashed, err := HashPassword(key)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.Contains(t, hashed, "$argon2id$")
	
	// Should be able to verify the key against the hash
	valid, err := VerifyPassword(key, hashed)
	assert.NoError(t, err)
	assert.True(t, valid)
}

