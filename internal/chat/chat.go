package chat

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatManager manages chat messages for rooms
type ChatManager struct {
	rooms map[string][]*Message
	mu    sync.RWMutex
}

// NewChatManager creates a new ChatManager instance
func NewChatManager() *ChatManager {
	return &ChatManager{
		rooms: make(map[string][]*Message),
	}
}

// AddMessage adds a new message to a room
func (cm *ChatManager) AddMessage(roomID, userID, username, content string) *Message {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Create message
	message := &Message{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		UserID:    userID,
		Username:  username,
		Content:   content,
		Timestamp: time.Now(),
	}
	
	// Add to room
	cm.rooms[roomID] = append(cm.rooms[roomID], message)
	
	// Keep only last 100 messages per room
	if len(cm.rooms[roomID]) > 100 {
		cm.rooms[roomID] = cm.rooms[roomID][1:]
	}
	
	return message
}

// GetMessages returns messages for a room
func (cm *ChatManager) GetMessages(roomID string) []*Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	// Return a copy of messages to prevent external modification
	messages := make([]*Message, len(cm.rooms[roomID]))
	copy(messages, cm.rooms[roomID])
	
	return messages
}

// GetRecentMessages returns the most recent messages for a room
func (cm *ChatManager) GetRecentMessages(roomID string, count int) []*Message {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	roomMessages := cm.rooms[roomID]
	
	// If count is greater than or equal to message count, return all messages
	if count >= len(roomMessages) {
		// Return a copy of messages to prevent external modification
		messages := make([]*Message, len(roomMessages))
		copy(messages, roomMessages)
		return messages
	}
	
	// Return the most recent messages
	startIndex := len(roomMessages) - count
	messages := make([]*Message, count)
	copy(messages, roomMessages[startIndex:])
	
	return messages
}

// DeleteMessagesForRoom deletes all messages for a room
func (cm *ChatManager) DeleteMessagesForRoom(roomID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	delete(cm.rooms, roomID)
}