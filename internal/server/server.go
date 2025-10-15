package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zubans/video-call-server/internal/auth"
	"github.com/zubans/video-call-server/internal/chat"
	"github.com/zubans/video-call-server/internal/metrics"
	"github.com/zubans/video-call-server/internal/models"
	"github.com/zubans/video-call-server/internal/recording"
	"github.com/zubans/video-call-server/internal/websocket"
)

// Server represents the video call server
type Server struct {
	router      *gin.Engine
	roomManager *models.RoomManager
	userManager *models.UserManager
	chatManager *chat.ChatManager
	recorder    *recording.Recorder
	hub         *websocket.Hub
	metrics     *metrics.Metrics
	httpServer  *http.Server
	wg          sync.WaitGroup
}

// NewServer creates a new Server instance
func NewServer() *Server {
	// Initialize room manager
	roomManager := &models.RoomManager{
		Rooms: make(map[string]*models.Room),
	}

	// Initialize user manager
	userManager := &models.UserManager{
		Users: make(map[string]*models.User),
	}

	// Initialize chat manager
	chatManager := chat.NewChatManager()

	// Initialize recorder
	recorder := recording.NewRecorder("./recordings")

	// Initialize WebSocket hub
	hub := websocket.NewHub()

	// Initialize metrics
	metr := metrics.AppMetrics

	return &Server{
		roomManager: roomManager,
		userManager: userManager,
		chatManager: chatManager,
		recorder:    recorder,
		hub:         hub,
		metrics:     metr,
	}
}

// Initialize sets up the server routes and components
func (s *Server) Initialize() {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	s.router = gin.Default()

	// Start WebSocket hub
	go s.hub.Run()

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8181"
	}

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: s.router,
	}
}

// setupRoutes sets up the server routes
func (s *Server) setupRoutes() {
	s.router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	// Public routes
	s.router.POST("/register", s.registerHandler)
	s.router.POST("/login", s.loginHandler)
	s.router.GET("/health", s.healthHandler)

	// Protected routes
	authorized := s.router.Group("/")
	authorized.Use(s.authMiddleware())
	{
		// Room management
		authorized.POST("/create-room", s.createRoomHandler)
		authorized.POST("/join-room", s.joinRoomHandler)
		authorized.POST("/leave-room", s.leaveRoomHandler)
		authorized.GET("/rooms", s.listRoomsHandler)

		// WebSocket connection
		authorized.GET("/ws", func(c *gin.Context) {
			websocket.ServeWs(s.hub, c.Writer, c.Request)
		})

		// Chat
		authorized.POST("/chat/send", s.sendChatMessageHandler)
		authorized.GET("/chat/history/:room_id", s.getChatHistoryHandler)

		// Recording
		authorized.POST("/recording/start", s.startRecordingHandler)
		authorized.POST("/recording/stop", s.stopRecordingHandler)
		authorized.GET("/recording/list/:room_id", s.listRecordingsHandler)

		// Metrics
		authorized.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}
}

// authMiddleware is a middleware for JWT authentication
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// Fallback: allow token via query param for WebSocket handshake (browsers can't set custom headers)
			tokenString = c.Query("token")
		}
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}

// Run starts the server
func (s *Server) Run() {
	// Initialize server
	s.Initialize()

	// Start HTTP server in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		log.Printf("Video call server starting on port %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
		log.Println("Server shutdown complete")
	}()

	// Wait for server to stop
	s.wg.Wait()
}

// registerHandler handles user registration
func (s *Server) registerHandler(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Register user
	user, err := auth.RegisterUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update metrics
	s.metrics.IncrementUsersRegistered()

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"user_id": user.ID,
	})
}

// loginHandler handles user login
func (s *Server) loginHandler(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier" binding:"required"` // username or email
		Password   string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, err := auth.AuthenticateUser(req.Identifier, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": user.ID,
	})
}

// createRoomHandler handles room creation
func (s *Server) createRoomHandler(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	_ = c.MustGet("username").(string)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create room
	s.roomManager.Mu.Lock()
	roomID := generateRoomID()
	room := &models.Room{
		ID:        roomID,
		Name:      req.Name,
		CreatorID: userID,
		Clients:   make(map[string]*models.Client),
		CreatedAt: time.Now(),
		IsActive:  true,
	}
	s.roomManager.Rooms[roomID] = room
	s.roomManager.Mu.Unlock()

	// Update metrics
	s.metrics.IncrementRoomsCreated()
	s.metrics.SetRoomsActive(float64(len(s.roomManager.Rooms)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Room created successfully",
		"room_id": room.ID,
		"name":    room.Name,
	})
}

// joinRoomHandler handles joining a room
func (s *Server) joinRoomHandler(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	username := c.MustGet("username").(string)

	var req struct {
		RoomID string `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find room
	s.roomManager.Mu.RLock()
	room, exists := s.roomManager.Rooms[req.RoomID]
	s.roomManager.Mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Create WebRTC peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create peer connection"})
		return
	}

	// Create client
	client := &models.Client{
		ID:       generateClientID(),
		UserID:   userID,
		Username: username,
		Conn:     peerConnection,
		Signal:   make(chan interface{}, 100),
		JoinedAt: time.Now(),
	}

	// Add client to room
	room.Mu.Lock()
	room.Clients[client.ID] = client
	room.Mu.Unlock()

	// Update metrics
	s.metrics.SetRoomParticipants(room.ID, float64(len(room.Clients)))

	// Setup WebRTC event handlers
	s.setupWebRTCEvents(room, client)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Joined room successfully",
		"room_id":   room.ID,
		"client_id": client.ID,
	})
}

// leaveRoomHandler handles leaving a room
func (s *Server) leaveRoomHandler(c *gin.Context) {
	_ = c.MustGet("user_id").(string)

	var req struct {
		RoomID   string `json:"room_id" binding:"required"`
		ClientID string `json:"client_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find room
	s.roomManager.Mu.RLock()
	room, exists := s.roomManager.Rooms[req.RoomID]
	s.roomManager.Mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Remove client from room
	room.Mu.Lock()
	client, clientExists := room.Clients[req.ClientID]
	if clientExists {
		// Close peer connection
		if client.Conn != nil {
			client.Conn.Close()
		}

		// Close signal channel
		close(client.Signal)

		// Remove client
		delete(room.Clients, req.ClientID)
	}
	room.Mu.Unlock()

	if !clientExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Update metrics
	s.metrics.SetRoomParticipants(room.ID, float64(len(room.Clients)))

	c.JSON(http.StatusOK, gin.H{
		"message": "Left room successfully",
	})
}

// sendChatMessageHandler handles sending chat messages
func (s *Server) sendChatMessageHandler(c *gin.Context) {
	userID := c.MustGet("user_id").(string)
	username := c.MustGet("username").(string)

	var req struct {
		RoomID  string `json:"room_id" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add message to chat
	message := s.chatManager.AddMessage(req.RoomID, userID, username, req.Message)

	// Update metrics
	s.metrics.IncrementChatMessagesSent()

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent successfully",
		"data":    message,
	})
}

// getChatHistoryHandler handles getting chat history
func (s *Server) getChatHistoryHandler(c *gin.Context) {
	roomID := c.Param("room_id")

	// Get messages
	messages := s.chatManager.GetRecentMessages(roomID, 50)

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

// startRecordingHandler handles starting a recording
func (s *Server) startRecordingHandler(c *gin.Context) {
	_ = c.MustGet("user_id").(string)

	var req struct {
		RoomID string `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start recording
	recording, err := s.recorder.StartRecording(req.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start recording"})
		return
	}

	// Update metrics
	s.metrics.IncrementRecordingsStarted()

	c.JSON(http.StatusOK, gin.H{
		"message":      "Recording started successfully",
		"recording_id": recording.ID,
	})
}

// stopRecordingHandler handles stopping a recording
func (s *Server) stopRecordingHandler(c *gin.Context) {
	var req struct {
		RecordingID string `json:"recording_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Stop recording
	err := s.recorder.StopRecording(req.RecordingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop recording"})
		return
	}

	// Update metrics
	s.metrics.IncrementRecordingsCompleted()

	c.JSON(http.StatusOK, gin.H{
		"message": "Recording stopped successfully",
	})
}

// listRecordingsHandler handles listing recordings for a room
func (s *Server) listRecordingsHandler(c *gin.Context) {
	roomID := c.Param("room_id")

	// List recordings
	recordings := s.recorder.ListRecordings(roomID)

	c.JSON(http.StatusOK, gin.H{
		"recordings": recordings,
	})
}

// healthHandler handles health checks
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Video call server is running",
	})
}

// setupWebRTCEvents sets up WebRTC event handlers
func (s *Server) setupWebRTCEvents(room *models.Room, client *models.Client) {
	// Handle ICE candidates
	client.Conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Broadcast ICE candidate to other clients
			room.Mu.RLock()
			for clientID, otherClient := range room.Clients {
				if clientID != client.ID && otherClient.Signal != nil {
					select {
					case otherClient.Signal <- models.SignalMessage{
						Type:      "ice-candidate",
						Data:      candidate.ToJSON(),
						Timestamp: time.Now(),
						SenderID:  client.ID,
					}:
					default:
						log.Printf("Signal channel full for client %s", clientID)
					}
				}
			}
			room.Mu.RUnlock()
		}
	})

	// Handle tracks
	client.Conn.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Log track reception
		log.Printf("Track received from client %s: %s", client.ID, track.Kind())
	})

	// Handle connection state changes
	client.Conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Printf("Connection state changed for client %s: %s", client.ID, state.String())

		// If connection is closed, remove client from room
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed {
			room.Mu.Lock()
			delete(room.Clients, client.ID)
			room.Mu.Unlock()

			// Update metrics
			s.metrics.SetRoomParticipants(room.ID, float64(len(room.Clients)))
		}
	})
}

// listRoomsHandler handles listing active rooms
func (s *Server) listRoomsHandler(c *gin.Context) {
	s.roomManager.Mu.RLock()
	defer s.roomManager.Mu.RUnlock()

	var rooms []gin.H
	for _, room := range s.roomManager.Rooms {
		room.Mu.RLock()
		rooms = append(rooms, gin.H{
			"id":                room.ID,
			"name":              room.Name,
			"creator_id":        room.CreatorID,
			"participant_count": len(room.Clients),
			"created_at":        room.CreatedAt,
			"is_active":         room.IsActive,
		})
		room.Mu.RUnlock()
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// generateRoomID generates a simple room ID (in production, use UUID)
func generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}

// generateClientID generates a simple client ID (in production, use UUID)
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
