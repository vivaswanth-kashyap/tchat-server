package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vivaswanth-kashyap/tchat-server/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MessageHandlers struct {
	DB *gorm.DB
}

func NewMessageHandlers(db *gorm.DB) *MessageHandlers {
	return &MessageHandlers{DB: db}
}

func (h *MessageHandlers) NewMessage(c *gin.Context) {
	var message models.Message
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err.Error())})
		return
	}
	result := h.DB.Clauses(clause.Returning{}).Create(&message)
	if result.Error != nil {
		log.Printf("Error creating new messaging: %v", result.Error)

		c.JSON(http.StatusBadRequest, gin.H{"error": "Error creating new message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"feedback": "New message created successfully", "message": message})
}
