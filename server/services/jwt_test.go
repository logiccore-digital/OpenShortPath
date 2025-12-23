package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"openshortpath/server/config"
)

// generateTestRSAKeyPair generates a test RSA key pair for testing
func generateTestRSAKeyPair(t *testing.T) (*rsa.PrivateKey, string, string) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKey, string(privateKeyPEM), string(publicKeyPEM)
}

func TestSignToken_HS256(t *testing.T) {
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret-key-123",
	}

	userID := "user123"
	token, err := SignToken(userID, cfg)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed and contains correct sub claim
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["sub"])
}

func TestSignToken_RS256(t *testing.T) {
	_, privateKeyPEM, publicKeyPEM := generateTestRSAKeyPair(t)

	cfg := &config.JWT{
		Algorithm:  "RS256",
		PrivateKey: privateKeyPEM,
		PublicKey:  publicKeyPEM,
	}

	userID := "user456"
	token, err := SignToken(userID, cfg)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token can be parsed and contains correct sub claim
	block, _ := pem.Decode([]byte(publicKeyPEM))
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	assert.NoError(t, err)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["sub"])
}

func TestSignToken_NilConfig(t *testing.T) {
	token, err := SignToken("user123", nil)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "JWT config not provided")
}

func TestSignToken_HS256_MissingSecretKey(t *testing.T) {
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "",
	}

	token, err := SignToken("user123", cfg)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "secret_key is required")
}

func TestSignToken_RS256_MissingPrivateKey(t *testing.T) {
	_, _, publicKeyPEM := generateTestRSAKeyPair(t)

	cfg := &config.JWT{
		Algorithm:  "RS256",
		PrivateKey: "",
		PublicKey:  publicKeyPEM,
	}

	token, err := SignToken("user123", cfg)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "private_key is required")
}

func TestSignToken_UnsupportedAlgorithm(t *testing.T) {
	cfg := &config.JWT{
		Algorithm: "ES256",
		SecretKey: "test",
	}

	token, err := SignToken("user123", cfg)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "unsupported algorithm")
}

func TestSignToken_RS256_InvalidPrivateKey(t *testing.T) {
	cfg := &config.JWT{
		Algorithm:  "RS256",
		PrivateKey: "invalid-pem-format",
		PublicKey:  "some-public-key",
	}

	token, err := SignToken("user123", cfg)
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to parse private key")
}

func TestSignToken_DifferentUserIDs(t *testing.T) {
	cfg := &config.JWT{
		Algorithm: "HS256",
		SecretKey: "test-secret",
	}

	userID1 := "user1"
	userID2 := "user2"

	token1, err1 := SignToken(userID1, cfg)
	token2, err2 := SignToken(userID2, cfg)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2)

	// Verify tokens contain correct user IDs
	parsed1, _ := jwt.Parse(token1, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})
	parsed2, _ := jwt.Parse(token2, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})

	claims1, _ := parsed1.Claims.(jwt.MapClaims)
	claims2, _ := parsed2.Claims.(jwt.MapClaims)

	assert.Equal(t, userID1, claims1["sub"])
	assert.Equal(t, userID2, claims2["sub"])
}

func TestSignToken_RS256_PKCS8Format(t *testing.T) {
	// Generate key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Encode as PKCS8
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	cfg := &config.JWT{
		Algorithm:  "RS256",
		PrivateKey: string(privateKeyPEM),
		PublicKey:  string(publicKeyPEM),
	}

	userID := "user789"
	token, err := SignToken(userID, cfg)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token
	block, _ := pem.Decode([]byte(publicKeyPEM))
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	assert.NoError(t, err)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims["sub"])
}
