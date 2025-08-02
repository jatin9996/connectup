package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/connect-up/auth-service/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper origin checking
	},
}

// WebSocketConnection represents a WebSocket connection
type WebSocketConnection struct {
	conn   *websocket.Conn
	userID string
	send   chan []byte
	mu     sync.Mutex
}

// WebSocketHandler handles WebSocket connections and messaging
type WebSocketHandler struct {
	connections map[string]*WebSocketConnection
	mu          sync.RWMutex
	kafkaWriter *kafka.Writer
	kafkaReader *kafka.Reader
	db          *models.DB
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(kafkaWriter *kafka.Writer, kafkaReader *kafka.Reader, db *models.DB) *WebSocketHandler {
	handler := &WebSocketHandler{
		connections: make(map[string]*WebSocketConnection),
		kafkaWriter: kafkaWriter,
		kafkaReader: kafkaReader,
		db:          db,
	}

	// Start Kafka consumer for chat messages
	go handler.startKafkaConsumer()

	return handler
}

// HandleWebSocket handles WebSocket connections
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create WebSocket connection
	wsConn := &WebSocketConnection{
		conn:   conn,
		userID: userID.(string),
		send:   make(chan []byte, 256),
	}

	// Register connection
	h.mu.Lock()
	h.connections[userID.(string)] = wsConn
	h.mu.Unlock()

	// Start goroutines for reading and writing
	go wsConn.writePump()
	go wsConn.readPump(h)

	// Send welcome message
	welcomeMsg := map[string]interface{}{
		"type":      "connection_established",
		"user_id":   userID.(string),
		"timestamp": time.Now().Unix(),
	}

	welcomeJSON, _ := json.Marshal(welcomeMsg)
	wsConn.send <- welcomeJSON
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *WebSocketConnection) readPump(h *WebSocketHandler) {
	defer func() {
		h.unregisterConnection(c.userID)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512) // Max message size
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Parse message
		var msgData map[string]interface{}
		if err := json.Unmarshal(message, &msgData); err != nil {
			log.Printf("Failed to parse message: %v", err)
			continue
		}

		// Handle different message types
		msgType, exists := msgData["type"].(string)
		if !exists {
			continue
		}

		switch msgType {
		case "chat_message":
			h.handleChatMessage(c.userID, msgData)
		case "typing":
			h.handleTypingEvent(c.userID, msgData)
		case "read_receipt":
			h.handleReadReceipt(c.userID, msgData)
		case "ping":
			// Send pong response
			pongMsg := map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now().Unix(),
			}
			pongJSON, _ := json.Marshal(pongMsg)
			c.send <- pongJSON
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleChatMessage handles incoming chat messages
func (h *WebSocketHandler) handleChatMessage(senderID string, msgData map[string]interface{}) {
	receiverID, exists := msgData["receiver_id"].(string)
	if !exists {
		return
	}

	content, exists := msgData["content"].(string)
	if !exists || content == "" {
		return
	}

	// Create message object
	message := models.Message{
		SenderID:    senderID,
		ReceiverID:  receiverID,
		Content:     content,
		MessageType: "text",
		IsRead:      false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save message to database
	if err := h.saveMessage(&message); err != nil {
		log.Printf("Failed to save message: %v", err)
		return
	}

	// Publish to Kafka
	h.publishChatMessage(&message)

	// Send to receiver if online
	h.sendToUser(receiverID, map[string]interface{}{
		"type":      "chat_message",
		"message":   message,
		"timestamp": time.Now().Unix(),
	})

	// Send confirmation to sender
	h.sendToUser(senderID, map[string]interface{}{
		"type":       "message_sent",
		"message_id": message.ID,
		"timestamp":  time.Now().Unix(),
	})
}

// handleTypingEvent handles typing indicators
func (h *WebSocketHandler) handleTypingEvent(userID string, msgData map[string]interface{}) {
	receiverID, exists := msgData["receiver_id"].(string)
	if !exists {
		return
	}

	isTyping, exists := msgData["is_typing"].(bool)
	if !exists {
		return
	}

	// Send typing indicator to receiver
	h.sendToUser(receiverID, map[string]interface{}{
		"type":      "typing_indicator",
		"user_id":   userID,
		"is_typing": isTyping,
		"timestamp": time.Now().Unix(),
	})
}

// handleReadReceipt handles read receipts
func (h *WebSocketHandler) handleReadReceipt(userID string, msgData map[string]interface{}) {
	messageID, exists := msgData["message_id"].(string)
	if !exists {
		return
	}

	// Update message as read in database
	if err := h.markMessageAsRead(messageID); err != nil {
		log.Printf("Failed to mark message as read: %v", err)
		return
	}

	// Send read receipt to sender
	h.sendToUser(userID, map[string]interface{}{
		"type":       "read_receipt",
		"message_id": messageID,
		"read_by":    userID,
		"timestamp":  time.Now().Unix(),
	})
}

// startKafkaConsumer starts consuming chat messages from Kafka
func (h *WebSocketHandler) startKafkaConsumer() {
	for {
		ctx := context.Background()
		m, err := h.kafkaReader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			continue
		}

		// Parse message
		var msgData map[string]interface{}
		if err := json.Unmarshal(m.Value, &msgData); err != nil {
			log.Printf("Failed to parse Kafka message: %v", err)
			continue
		}

		// Handle different message types
		msgType, exists := msgData["type"].(string)
		if !exists {
			continue
		}

		switch msgType {
		case "chat_message":
			h.broadcastChatMessage(msgData)
		case "user_status":
			h.broadcastUserStatus(msgData)
		}
	}
}

// publishChatMessage publishes a chat message to Kafka
func (h *WebSocketHandler) publishChatMessage(message *models.Message) {
	if h.kafkaWriter == nil {
		return
	}

	msgData := map[string]interface{}{
		"type":      "chat_message",
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	msgJSON, err := json.Marshal(msgData)
	if err != nil {
		return
	}

	h.kafkaWriter.WriteMessages(context.Background(), kafka.Message{
		Topic: "chat-messages",
		Key:   []byte(message.SenderID),
		Value: msgJSON,
	})
}

// broadcastChatMessage broadcasts a chat message to relevant users
func (h *WebSocketHandler) broadcastChatMessage(msgData map[string]interface{}) {
	message, exists := msgData["message"].(map[string]interface{})
	if !exists {
		return
	}

	receiverID, exists := message["receiver_id"].(string)
	if !exists {
		return
	}

	// Send to receiver
	h.sendToUser(receiverID, msgData)
}

// broadcastUserStatus broadcasts user status changes
func (h *WebSocketHandler) broadcastUserStatus(msgData map[string]interface{}) {
	userID, exists := msgData["user_id"].(string)
	if !exists {
		return
	}

	// Broadcast to all connected users (or implement more sophisticated logic)
	h.mu.RLock()
	for _, conn := range h.connections {
		if conn.userID != userID {
			conn.send <- []byte(fmt.Sprintf(`{"type":"user_status","user_id":"%s","status":"%s"}`,
				userID, msgData["status"]))
		}
	}
	h.mu.RUnlock()
}

// sendToUser sends a message to a specific user
func (h *WebSocketHandler) sendToUser(userID string, message map[string]interface{}) {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return
	}

	conn.send <- messageJSON
}

// unregisterConnection removes a connection from the handler
func (h *WebSocketHandler) unregisterConnection(userID string) {
	h.mu.Lock()
	delete(h.connections, userID)
	h.mu.Unlock()

	// Broadcast user offline status
	h.broadcastUserStatus(map[string]interface{}{
		"user_id": userID,
		"status":  "offline",
	})
}

// saveMessage saves a message to the database
func (h *WebSocketHandler) saveMessage(message *models.Message) error {
	query := `
		INSERT INTO messages (sender_id, receiver_id, content, message_type, is_read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	return h.db.QueryRow(query,
		message.SenderID, message.ReceiverID, message.Content, message.MessageType,
		message.IsRead, message.CreatedAt, message.UpdatedAt,
	).Scan(&message.ID)
}

// markMessageAsRead marks a message as read
func (h *WebSocketHandler) markMessageAsRead(messageID string) error {
	query := `
		UPDATE messages SET is_read = true, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := h.db.Exec(query, messageID)
	return err
}

// GetOnlineUsers returns a list of online users
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	h.mu.RLock()
	onlineUsers := make([]string, 0, len(h.connections))
	for userID := range h.connections {
		onlineUsers = append(onlineUsers, userID)
	}
	h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"online_users": onlineUsers,
		"count":        len(onlineUsers),
	})
}
