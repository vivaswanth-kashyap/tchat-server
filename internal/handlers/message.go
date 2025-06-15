package handlers

import (
	"errors"
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

func (h *MessageHandlers) ReadMessage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing message id"})
		return
	}

	var msg models.Message
	err := h.DB.First(&msg, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (h *MessageHandlers) ReadChat(c *gin.Context) {
	senderId := c.Query("sender_id")
	receiverId := c.Query("receiver_id")

	if senderId == "" || receiverId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderid and/or receiverid missing"})
	}

	var messages []models.Message

	result := h.DB.Where("sender_id = ? AND receiver_id = ?", senderId, receiverId).
		Find(&messages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandlers) ReadLastSent(c *gin.Context) {
	senderId := c.Query("sender_id")
	receiverId := c.Query("receiver_id")

	if senderId == "" || receiverId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderid and/or receiverid missing"})
	}

	var message models.Message

	result := h.DB.Where("sender_id = ? AND receiver_id = ?", senderId, receiverId).
		Order("created_at DESC").
		First(&message)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No messages found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})

}
