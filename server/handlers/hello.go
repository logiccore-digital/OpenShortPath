package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HelloHandler struct {
	db *gorm.DB
}

func NewHelloHandler(db *gorm.DB) *HelloHandler {
	return &HelloHandler{
		db: db,
	}
}

func (h *HelloHandler) HelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

