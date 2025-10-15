package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// User представляет пользователя системы
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Room представляет собой комнату для видеозвонка
type Room struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	CreatorID   string             `json:"creator_id"`
	Clients     map[string]*Client `json:"clients"`
	ChatHistory []ChatMessage      `json:"chat_history"`
	CreatedAt   time.Time          `json:"created_at"`
	IsActive    bool               `json:"is_active"`
	Mu          sync.RWMutex
}

// Client представляет собой клиента в комнате
type Client struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Conn        *webrtc.PeerConnection `json:"-"` // Не сериализуем в JSON
	WebSocket   *WebSocketConnection   `json:"-"`
	Signal      chan interface{}       `json:"-"`
	JoinedAt    time.Time              `json:"joined_at"`
	IsRecording bool                   `json:"is_recording"`
	RecordingID string                 `json:"recording_id,omitempty"`
}

// WebSocketConnection представляет WebSocket соединение клиента
type WebSocketConnection struct {
	Conn       *websocket.Conn
	Send       chan []byte
	ClientID   string
	RoomID     string
	LastActive time.Time
}

// SignalMessage представляет сообщение сигнализации
type SignalMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	SenderID  string      `json:"sender_id,omitempty"`
}

// ChatMessage представляет сообщение в чате
type ChatMessage struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Recording представляет запись звонка
type Recording struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Filename  string    `json:"filename"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at,omitempty"`
	Duration  int       `json:"duration,omitempty"`
	Size      int64     `json:"size,omitempty"`
	Status    string    `json:"status"` // active, completed, failed
}

// AuthToken представляет токен аутентификации
type AuthToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// RoomManager управляет всеми комнатами
type RoomManager struct {
	Rooms map[string]*Room
	Mu    sync.RWMutex
}

// UserManager управляет пользователями
type UserManager struct {
	Users map[string]*User
	mu    sync.RWMutex
}

// RecordingManager управляет записями звонков
type RecordingManager struct {
	Recordings map[string]*Recording
	mu         sync.RWMutex
}

// ChatManager управляет сообщениями чата
type ChatManager struct {
	Messages map[string][]ChatMessage
	mu       sync.RWMutex
}
