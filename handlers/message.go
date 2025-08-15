package handlers

import (
	"net/http"
	"strconv"
	"time"

	"ecommerce-app/models"
	"ecommerce-app/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db               *gorm.DB
	websocketService *services.WebSocketService
}

func NewMessageHandler(db *gorm.DB, websocketService *services.WebSocketService) *MessageHandler {
	return &MessageHandler{
		db:               db,
		websocketService: websocketService,
	}
}

type SendMessageRequest struct {
	ToUserID uint   `json:"to_user_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	fromUserID := c.MustGet("user_id").(uint)

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if recipient exists
	var toUser models.User
	if err := h.db.First(&toUser, req.ToUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recipient"})
		return
	}

	// Don't allow sending message to yourself
	if fromUserID == req.ToUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send message to yourself"})
		return
	}

	// Create message
	message := models.Message{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		Content:    req.Content,
		IsRead:     false,
	}

	if err := h.db.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Send real-time message via WebSocket
	h.websocketService.SendPrivateMessage(fromUserID, req.ToUserID, req.Content)

	c.JSON(http.StatusCreated, gin.H{"message": message})
}

func (h *MessageHandler) GetConversation(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	otherUserID := c.Param("user_id")
	otherID, err := strconv.ParseUint(otherUserID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if other user exists
	var otherUser models.User
	if err := h.db.First(&otherUser, otherID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	// Get messages between the two users
	var messages []models.Message
	if err := h.db.Preload("FromUser", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name")
	}).Preload("ToUser", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, first_name, last_name")
	}).Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		userID, otherID, otherID, userID).Order("created_at ASC").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Mark messages as read
	if err := h.db.Model(&models.Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", otherID, userID, false).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark messages as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *MessageHandler) GetConversations(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Get all conversations (users with whom the current user has exchanged messages)
	var conversations []gin.H
	rows, err := h.db.Raw(`
		SELECT DISTINCT 
			u.id, u.first_name, u.last_name, u.email,
			(SELECT content FROM messages 
			 WHERE ((from_user_id = ? AND to_user_id = u.id) OR (from_user_id = u.id AND to_user_id = ?))
			 ORDER BY created_at DESC LIMIT 1) as last_message,
			(SELECT created_at FROM messages 
			 WHERE ((from_user_id = ? AND to_user_id = u.id) OR (from_user_id = u.id AND to_user_id = ?))
			 ORDER BY created_at DESC LIMIT 1) as last_message_time,
			(SELECT COUNT(*) FROM messages 
			 WHERE from_user_id = u.id AND to_user_id = ? AND is_read = false) as unread_count
		FROM users u
		WHERE u.id IN (
			SELECT DISTINCT 
				CASE 
					WHEN from_user_id = ? THEN to_user_id 
					ELSE from_user_id 
				END
			FROM messages 
			WHERE from_user_id = ? OR to_user_id = ?
		)
		ORDER BY last_message_time DESC
	`, userID, userID, userID, userID, userID, userID, userID, userID).Rows()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		var firstName, lastName, email, lastMessage string
		var lastMessageTime time.Time
		var unreadCount int

		if err := rows.Scan(&id, &firstName, &lastName, &email, &lastMessage, &lastMessageTime, &unreadCount); err != nil {
			continue
		}

		conversations = append(conversations, gin.H{
			"user_id":           id,
			"first_name":        firstName,
			"last_name":         lastName,
			"email":             email,
			"last_message":      lastMessage,
			"last_message_time": lastMessageTime,
			"unread_count":      unreadCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func (h *MessageHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var count int64
	if err := h.db.Model(&models.Message{}).Where("to_user_id = ? AND is_read = ?", userID, false).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	messageID := c.Param("id")
	id, err := strconv.ParseUint(messageID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.db.Model(&models.Message{}).Where("id = ? AND to_user_id = ?", id, userID).Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	messageID := c.Param("id")
	id, err := strconv.ParseUint(messageID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := h.db.Where("id = ? AND from_user_id = ?", id, userID).Delete(&models.Message{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}
