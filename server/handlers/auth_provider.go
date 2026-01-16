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
	response := gin.H{
		"auth_provider": h.cfg.AuthProvider,
		"enable_signup": h.cfg.EnableSignup,
	}

	// Include Clerk publishable key if auth provider is Clerk
	if h.cfg.AuthProvider == "clerk" && h.cfg.Clerk != nil {
		response["clerk_publishable_key"] = h.cfg.Clerk.PublishableKey
	}

	c.JSON(http.StatusOK, response)
}
