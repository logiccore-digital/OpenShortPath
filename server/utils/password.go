package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
)

// HashPassword hashes a password using argon2id
// Returns a string in the format: $argon2id$v=<version>$m=<memory>,t=<time>,p=<threads>$<salt>$<hash>
func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash the password
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Return formatted string
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword verifies a password against a hashed password
func VerifyPassword(password, hashedPassword string) (bool, error) {
	// Parse the hashed password
	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported hash algorithm: %s", parts[1])
	}

	// Parse version
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("failed to parse version: %w", err)
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version: %d", version)
	}

	// Parse parameters
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Compute hash of the provided password
	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	// Compare hashes using constant-time comparison
	if subtle.ConstantTimeCompare(hash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}

