package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"openshortpath/server/config"
)

type DomainsHandler struct {
	cfg *config.Config
}

type DomainsResponse struct {
	Domains []string `json:"domains"`
}

func NewDomainsHandler(cfg *config.Config) *DomainsHandler {
	return &DomainsHandler{
		cfg: cfg,
	}
}

// GetDomains returns the list of available short domains from the configuration
func (h *DomainsHandler) GetDomains(c *gin.Context) {
	response := DomainsResponse{
		Domains: h.cfg.AvailableShortDomains,
	}
	c.JSON(http.StatusOK, response)
}

