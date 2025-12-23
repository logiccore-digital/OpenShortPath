package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	// APIKeyPrefix is the prefix for all API keys
	APIKeyPrefix = "osp_sk_"
	// APIKeyRandomBytes is the number of random bytes to generate for the key
	APIKeyRandomBytes = 32
)

// GenerateAPIKey generates a new API key with the osp_sk_ prefix
// Returns a key in the format: osp_sk_<base64url-encoded-random-bytes>
func GenerateAPIKey() (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, APIKeyRandomBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64url (URL-safe base64 without padding)
	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	// Combine prefix with encoded bytes
	key := APIKeyPrefix + encoded

	return key, nil
}

