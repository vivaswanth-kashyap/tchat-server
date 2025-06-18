package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type MessageRequest struct {
	SenderID         string `json:"sender_id"`
	ReceiverID       string `json:"receiver_id,omitempty"`
	ReceiverUsername string `json:"receiver_username,omitempty"`
	Body             string `json:"body"`
}

func (h *MessageHandlers) NewMessage(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err.Error())})
	}

	receiverID := req.ReceiverID
	if req.ReceiverUsername != "" {
		var user models.User
		if err := h.DB.Where("username = ?", req.ReceiverUsername).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User Not Found"})
			return
		}
		receiverID = user.ID.String()
	}

	if receiverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiver_id or receiver_username required"})
		return
	}

	senderUUID, err := uuid.Parse(req.SenderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sender_id format"})
	}

	receiverUUID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invaliud receiver_id format"})
	}

	message := models.Message{
		SenderID:   senderUUID,
		ReceiverID: receiverUUID,
		Body:       req.Body,
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

	if senderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderid missing"})
	}

	receiverId, err := h.resolveReceiverID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

	if senderId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "senderid missing"})
	}

	receiverId, err := h.resolveReceiverID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

func (h *MessageHandlers) resolveReceiverID(c *gin.Context) (string, error) {
	receiverId := c.Query("receiver_id")
	receiverUsername := c.Query("receiver_username")

	if receiverUsername == "" {
		return "", fmt.Errorf("receiver id or username required")
	}

	if receiverUsername != "" {
		var user models.User
		if err := h.DB.Where("username = ?", receiverUsername).First(&user).Error; err != nil {
			return "", fmt.Errorf("user not found")
		}
		return user.ID.String(), nil
	}
	return receiverId, nil
}
