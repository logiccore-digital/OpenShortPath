package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"openshortpath/server/config"

	"github.com/golang-jwt/jwt/v5"
)

// SignToken signs a JWT token with the user ID as the 'sub' claim
func SignToken(userID string, jwtConfig *config.JWT) (string, error) {
	if jwtConfig == nil {
		return "", fmt.Errorf("JWT config not provided")
	}

	// Create claims with user ID as 'sub'
	claims := jwt.MapClaims{
		"sub": userID,
	}

	// Sign token based on algorithm
	switch jwtConfig.Algorithm {
	case "HS256":
		if jwtConfig.SecretKey == "" {
			return "", fmt.Errorf("secret_key is required for HS256")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString([]byte(jwtConfig.SecretKey))

	case "RS256":
		if jwtConfig.PrivateKey == "" {
			return "", fmt.Errorf("private_key is required for RS256")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		privateKey, err := parseRSAPrivateKey(jwtConfig.PrivateKey)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
		return token.SignedString(privateKey)

	default:
		return "", fmt.Errorf("unsupported algorithm: %s", jwtConfig.Algorithm)
	}
}

// parseRSAPrivateKey parses a PEM-encoded RSA private key
func parseRSAPrivateKey(pemString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS1 format first
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return privateKey, nil
	}

	// Try PKCS8 format
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaKey, nil
}

