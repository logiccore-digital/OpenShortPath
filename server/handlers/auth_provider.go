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
	c.JSON(http.StatusOK, gin.H{
		"auth_provider": h.cfg.AuthProvider,
		"enable_signup": h.cfg.EnableSignup,
	})
}
