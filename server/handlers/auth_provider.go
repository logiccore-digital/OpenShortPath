package handlers

import (
	"net/http"

	"openshortpath/server/config"

	"github.com/gin-gonic/gin"
)

type AuthProviderHandler struct {
	cfg *config.Config
}

func NewAuthProviderHandler(cfg *config.Config) *AuthProviderHandler {
	return &AuthProviderHandler{
		cfg: cfg,
	}
}

func (h *AuthProviderHandler) GetAuthProvider(c *gin.Context) {
	authProvider := h.cfg.AuthProvider
	// Default to "external_jwt" if not set (as per config comments)
	if authProvider == "" {
		authProvider = "external_jwt"
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_provider": authProvider,
	})
}

