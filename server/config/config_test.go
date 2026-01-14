package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate_LocalAuth_HS256(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_LocalAuth_RS256(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm:  "RS256",
			PrivateKey: "test-private-key",
			PublicKey:  "test-public-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_LocalAuth_MissingJWT(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT:          nil,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT config is required")
}

func TestConfig_Validate_LocalAuth_HS256_MissingSecretKey(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret_key is required")
}

func TestConfig_Validate_LocalAuth_RS256_MissingPrivateKey(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm:  "RS256",
			PrivateKey: "",
			PublicKey:  "test-public-key",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "private_key is required")
}

func TestConfig_Validate_LocalAuth_RS256_MissingPublicKey(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm:  "RS256",
			PrivateKey: "test-private-key",
			PublicKey:  "",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public_key is required")
}

func TestConfig_Validate_ExternalJWT_HS256(t *testing.T) {
	cfg := &Config{
		AuthProvider: "external_jwt",
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_ExternalJWT_RS256(t *testing.T) {
	cfg := &Config{
		AuthProvider: "external_jwt",
		JWT: &JWT{
			Algorithm: "RS256",
			PublicKey: "test-public-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_ExternalJWT_RS256_MissingPublicKey(t *testing.T) {
	cfg := &Config{
		AuthProvider: "external_jwt",
		JWT: &JWT{
			Algorithm: "RS256",
			PublicKey: "",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "public_key is required")
}

func TestConfig_Validate_UnsupportedAlgorithm(t *testing.T) {
	cfg := &Config{
		AuthProvider: "local",
		JWT: &JWT{
			Algorithm: "ES256",
			SecretKey: "test",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported JWT algorithm")
}

func TestConfig_Validate_NoAuthProvider(t *testing.T) {
	// When auth_provider is empty or not set, validation should fail
	cfg := &Config{
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth_provider is required")
}

func TestConfig_Validate_InvalidAuthProvider(t *testing.T) {
	// When auth_provider has invalid value, validation should fail
	cfg := &Config{
		AuthProvider: "invalid",
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth_provider")
}

func TestConfig_Validate_NoJWTConfig(t *testing.T) {
	// When JWT config is not provided and auth_provider is not local, should pass
	cfg := &Config{
		AuthProvider: "external_jwt",
		JWT:          nil,
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_Validate_EmptyAlgorithm(t *testing.T) {
	// When algorithm is empty but auth_provider is set, should pass
	cfg := &Config{
		AuthProvider: "external_jwt",
		JWT: &JWT{
			Algorithm: "",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfig_EnableSignup_Default(t *testing.T) {
	// Test that EnableSignup defaults to false
	cfg, err := LoadConfig("")
	assert.NoError(t, err)
	assert.False(t, cfg.EnableSignup)
}

func TestConfig_EnableSignup_True(t *testing.T) {
	// Test that EnableSignup can be set to true
	cfg := &Config{
		AuthProvider: "local",
		EnableSignup: true,
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
	assert.True(t, cfg.EnableSignup)
}

func TestConfig_EnableSignup_False(t *testing.T) {
	// Test that EnableSignup can be set to false explicitly
	cfg := &Config{
		AuthProvider: "local",
		EnableSignup: false,
		JWT: &JWT{
			Algorithm: "HS256",
			SecretKey: "test-secret-key",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
	assert.False(t, cfg.EnableSignup)
}
