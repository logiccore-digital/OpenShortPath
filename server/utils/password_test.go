package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "test-password-123"

	hashed, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.Contains(t, hashed, "$argon2id$")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hashed, err := HashPassword("")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashed)
}

func TestHashPassword_DifferentPasswords(t *testing.T) {
	password1 := "password1"
	password2 := "password2"

	hashed1, err1 := HashPassword(password1)
	hashed2, err2 := HashPassword(password2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hashed1, hashed2)
}

func TestHashPassword_UniqueHashes(t *testing.T) {
	password := "same-password"

	hashed1, err1 := HashPassword(password)
	hashed2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	// Hashes should be different due to random salt
	assert.NotEqual(t, hashed1, hashed2)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "test-password-123"
	hashed, err := HashPassword(password)
	assert.NoError(t, err)

	valid, err := VerifyPassword(password, hashed)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "test-password-123"
	wrongPassword := "wrong-password"
	hashed, err := HashPassword(password)
	assert.NoError(t, err)

	valid, err := VerifyPassword(wrongPassword, hashed)
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	password := "test-password"
	hashed, err := HashPassword(password)
	assert.NoError(t, err)

	valid, err := VerifyPassword("", hashed)
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	testCases := []struct {
		name string
		hash string
	}{
		{"Empty hash", ""},
		{"Invalid format", "not-a-hash"},
		{"Missing parts", "$argon2id$v=19"},
		{"Wrong algorithm", "$bcrypt$v=19$m=65536,t=1,p=4$salt$hash"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := VerifyPassword("password", tc.hash)
			assert.Error(t, err)
			assert.False(t, valid)
		})
	}
}

func TestVerifyPassword_InvalidVersion(t *testing.T) {
	// Create a hash with invalid version
	invalidHash := "$argon2id$v=99$m=65536,t=1,p=4$c2FsdA==$aGFzaA=="

	valid, err := VerifyPassword("password", invalidHash)
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestVerifyPassword_RoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"complex-P@ssw0rd!123",
		"very-long-password-with-many-characters-and-special-symbols-!@#$%^&*()",
		"1234567890",
		"   password with spaces   ",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hashed, err := HashPassword(password)
			assert.NoError(t, err)

			valid, err := VerifyPassword(password, hashed)
			assert.NoError(t, err)
			assert.True(t, valid, "Password verification failed for: %s", password)
		})
	}
}
