package websocket

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/pkg/structures"
)

// Manager handles all WebSocket connections and operations
type Manager struct {
	clients      map[string]*Client          // userID -> client
	connections  map[*websocket.Conn]*Client // conn -> client (for cleanup)
	clientsMutex sync.RWMutex

	// Configuration
	maxConnections    int
	heartbeatInterval time.Duration
	connectionTimeout time.Duration
	heartbeatTimeout  time.Duration
	readTimeout       time.Duration

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Services
	authService auth.Authmen

	// Server info for hello messages
	serverID string
	features []string
}

// Client represents a connected WebSocket client
type Client struct {
	Conn        *websocket.Conn
	User        *auth.JWTClaimUser
	UserID      string
	ConnectedAt time.Time
	LastPing    time.Time
	LastPong    time.Time

	// Channel for sending messages
	sendChan chan []byte

	// Context for this client's lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread-safe operations
	mu        sync.Mutex
	closeOnce sync.Once

	// Heartbeat tracking
	awaitingPong bool
}

// NewManager creates a new WebSocket manager
func NewManager(authService auth.Authmen) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		clients:           make(map[string]*Client),
		connections:       make(map[*websocket.Conn]*Client),
		maxConnections:    1000, // Configurable
		heartbeatInterval: 45 * time.Second,
		connectionTimeout: 5 * time.Minute,
		heartbeatTimeout:  15 * time.Second,
		readTimeout:       2 * time.Minute,
		ctx:               ctx,
		cancel:            cancel,
		authService:       authService,
		serverID:          "serra-ws-server",
		features:          []string{"heartbeat", "batch_downloads", "system_status"},
	}
}

// RegisterRoutes sets up the websocket endpoint
func (m *Manager) RegisterRoutes(gctx global.Context, router fiber.Router) {
	slog.Info("Registering WebSocket routes", "path", "/ws")
	
	router.Use("/ws", func(c *fiber.Ctx) error {
		slog.Info("🔌 WebSocket middleware triggered", 
			"path", c.Path(),
			"method", c.Method(),
			"remoteAddr", c.IP(),
			"userAgent", c.Get("User-Agent"),
			"origin", c.Get("Origin"),
			"upgrade", c.Get("Upgrade"),
			"connection", c.Get("Connection"),
			"serra_token", c.Cookies("serra_token"))
		
		if websocket.IsWebSocketUpgrade(c) {
			slog.Info("✅ Valid WebSocket upgrade request detected")
			return c.Next()
		}
		
		slog.Error("❌ Invalid WebSocket upgrade request", 
			"path", c.Path(),
			"method", c.Method(),
			"upgrade", c.Get("Upgrade"),
			"connection", c.Get("Connection"))
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		slog.Info("WebSocket connection established, entering handleConnection")
		m.handleConnection(c, gctx)
	}))
	
	slog.Info("WebSocket routes registered successfully")
}

// handleConnection manages a new WebSocket connection
func (m *Manager) handleConnection(c *websocket.Conn, gctx global.Context) {
	slog.Info("🔌 WebSocket handleConnection called",
		"remoteAddr", c.RemoteAddr().String(),
		"currentConnections", m.getConnectionCount())

	// Check connection limit
	if m.getConnectionCount() >= m.maxConnections {
		slog.Warn("Connection limit reached", "limit", m.maxConnections)
		m.sendErrorAndClose(c, "Server at capacity")
		return
	}

	// Extract and validate token
	slog.Debug("🔐 Starting authentication process", "remoteAddr", c.RemoteAddr().String())
	client, err := m.authenticateConnection(c)
	if err != nil {
		slog.Warn("❌ Authentication failed", "error", err, "remoteAddr", c.RemoteAddr().String())
		m.sendErrorAndClose(c, err.Error())
		return
	}
	slog.Info("✅ Authentication successful", "userID", client.UserID, "username", client.User.Username)

	// Register client
	slog.Debug("📝 Registering client", "userID", client.UserID)
	if !m.registerClient(client) {
		slog.Warn("Failed to register client", "userID", client.UserID)
		m.sendErrorAndClose(c, "Failed to register connection")
		return
	}
	slog.Info("✅ Client registered successfully", "userID", client.UserID, "totalClients", m.getConnectionCount())
	// defer m.unregisterClient(client)

	// Send hello message
	helloPayload := structures.HelloPayload{
		Message:  "Connected successfully to Serra WebSocket server",
		ServerID: m.serverID,
		Features: m.features,
		Metadata: map[string]string{
			"version":            "1.0",
			"heartbeat_interval": m.heartbeatInterval.String(),
			"timeout":            m.connectionTimeout.String(),
		},
	}

	err = m.sendMessage(client, structures.OpcodeHello, helloPayload)
	if err != nil {
		slog.Error("Failed to send hello message", "userID", client.UserID, "error", err)
		m.unregisterClient(client)
		return
	}

	slog.Info("WebSocket connected successfully",
		"userID", client.UserID,
		"username", client.User.Username,
		"totalConnections", m.getConnectionCount())

	// Start client message handling
	m.handleClientMessages(client)
}

// authenticateConnection validates the connection and creates a client
func (m *Manager) authenticateConnection(c *websocket.Conn) (*Client, error) {
	// Extract token from cookie
	slog.Debug("🔍 Checking for serra_token cookie")
	token := c.Cookies("serra_token")
	if token == "" {
		slog.Warn("❌ No serra_token cookie found")
		return nil, &AuthError{Message: "Missing auth token"}
	}
	slog.Debug("✅ Found serra_token cookie", "tokenLength", len(token))

	// Validate token
	claims, err := m.authService.ValidateJWT(token)
	if err != nil {
		return nil, &AuthError{Message: "Invalid auth token"}
	}

	// Create client context
	ctx, cancel := context.WithCancel(m.ctx)

	now := time.Now()
	client := &Client{
		Conn:         c,
		User:         claims,
		UserID:       claims.UserID,
		ConnectedAt:  now,
		LastPing:     now,
		LastPong:     now,
		sendChan:     make(chan []byte, 100), // Buffer for 100 messages
		ctx:          ctx,
		cancel:       cancel,
		awaitingPong: false,
	}

	return client, nil
}

// registerClient adds a client to the manager
func (m *Manager) registerClient(client *Client) bool {
	slog.Debug("🔒 Acquiring lock for client registration", "userID", client.UserID)
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if user already has a connection
	if existing, exists := m.clients[client.UserID]; exists {
		slog.Info("Replacing existing connection", "userID", client.UserID)
		m.closeClient(existing)
	}

	slog.Debug("📝 Adding client to maps", "userID", client.UserID)
	m.clients[client.UserID] = client
	m.connections[client.Conn] = client

	slog.Info("✅ Client added to maps", 
		"userID", client.UserID,
		"totalClients", len(m.clients),
		"totalConnections", len(m.connections))

	// Start background goroutines for this client
	slog.Debug("🚀 Starting background goroutines for client", "userID", client.UserID)
	go m.clientWriter(client)
	go m.clientHeartbeat(client)

	return true
}

// unregisterClient removes a client from the manager
func (m *Manager) unregisterClient(client *Client) {
	slog.Debug("🔒 Acquiring lock for client unregistration", "userID", client.UserID)
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	slog.Debug("🗑️ Removing client from maps", "userID", client.UserID)
	delete(m.clients, client.UserID)
	delete(m.connections, client.Conn)

	client.cancel()
	client.Conn.Close()
	client.closeOnce.Do(func() {
		close(client.sendChan)
	})

	slog.Info("❌ WebSocket disconnected", 
		"userID", client.UserID,
		"remainingClients", len(m.clients),
		"remainingConnections", len(m.connections))
}

// handleClientMessages processes incoming messages from a client
func (m *Manager) handleClientMessages(client *Client) {
	defer m.unregisterClient(client)

	// Set read deadline
	client.Conn.SetReadDeadline(time.Now().Add(m.readTimeout))

	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			_, message, err := client.Conn.ReadMessage()
			if err != nil {
				slog.Debug("Client read error", "userID", client.UserID, "error", err)
				return
			}

			// Handle message
			if err := m.handleMessage(client, message); err != nil {
				slog.Error("Failed to handle message", "userID", client.UserID, "error", err)
			}

			// Reset read deadline
			client.Conn.SetReadDeadline(time.Now().Add(m.readTimeout))
		}
	}
}

// handleMessage processes a single message from a client
func (m *Manager) handleMessage(client *Client, message []byte) error {
	slog.Debug("📨 Received message from client", "userID", client.UserID, "rawMessage", string(message))
	
	msg, err := structures.ParseMessage(message)
	if err != nil {
		slog.Warn("Invalid message received", "userID", client.UserID, "error", err, "rawMessage", string(message))
		return m.sendError(client, "Invalid message format")
	}

	slog.Debug("📦 Parsed message from client", "userID", client.UserID, "opcode", msg.Op, "timestamp", msg.Timestamp, "data", msg.Data)

	switch msg.Op {
	case structures.OpcodeHeartbeat:
		// Client sent us a heartbeat response - they're alive
		slog.Debug("💗 Received heartbeat response from client", "userID", client.UserID)
		client.mu.Lock()
		client.LastPing = time.Now()
		client.awaitingPong = false // Client responded to our ping
		client.mu.Unlock()

		// Don't send response - this would create infinite loop
		return nil

	default:
		slog.Debug("Unknown opcode", "userID", client.UserID, "opcode", msg.Op)
		return m.sendError(client, "Unknown operation")
	}
}

// clientWriter handles writing messages to a client
func (m *Manager) clientWriter(client *Client) {
	for {
		select {
		case <-client.ctx.Done():
			return
		case message := <-client.sendChan:
			client.mu.Lock()
			err := client.Conn.WriteMessage(websocket.TextMessage, message)
			client.mu.Unlock()

			if err != nil {
				slog.Error("Failed to write message", "userID", client.UserID, "error", err)
				return
			}
		}
	}
}

// clientHeartbeat sends periodic heartbeats and checks for stale connections
func (m *Manager) clientHeartbeat(client *Client) {
	ticker := time.NewTicker(m.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-client.ctx.Done():
			return
		case <-ticker.C:
			client.mu.Lock()
			lastPing := client.LastPing
			lastPong := client.LastPong
			awaitingPong := client.awaitingPong
			client.mu.Unlock()

			// Check if we're waiting for a pong response too long
			if awaitingPong && time.Since(lastPong) > m.heartbeatTimeout {
				slog.Warn("Client heartbeat timeout (no pong received)", "userID", client.UserID, "lastPong", lastPong, "timeout", m.heartbeatTimeout)
				return
			}

			// Check if client is generally unresponsive
			if time.Since(lastPing) > m.connectionTimeout {
				slog.Warn("Client connection timeout", "userID", client.UserID)
				return
			}

			// Only send heartbeat if we're not already waiting for a pong
			if !awaitingPong {
				slog.Debug("💓 Sending heartbeat to client", "userID", client.UserID, "awaitingPong", awaitingPong)
				// Send heartbeat ping
				now := time.Now()
				client.mu.Lock()
				client.awaitingPong = true
				client.LastPong = now  // Track when we sent this ping
				client.mu.Unlock()

				if err := m.sendMessage(client, structures.OpcodeHeartbeat, nil); err != nil {
					slog.Error("Failed to send heartbeat", "userID", client.UserID, "error", err)
					return
				}
			} else {
				slog.Debug("⏳ Skipping heartbeat - already awaiting pong", "userID", client.UserID, "awaitingPong", awaitingPong)
			}
		}
	}
}

// sendMessage sends a message to a specific client
func (m *Manager) sendMessage(client *Client, op structures.Opcode, data interface{}) error {
	msg := structures.NewMessage(op, data)
	payload, err := structures.MarshalMessage(msg)
	if err != nil {
		return err
	}

	select {
	case client.sendChan <- payload:
		return nil
	case <-client.ctx.Done():
		return &ClientError{Message: "Client disconnected"}
	default:
		return &ClientError{Message: "Client buffer full"}
	}
}

// sendError sends an error message to a client
func (m *Manager) sendError(client *Client, message string) error {
	return m.sendMessage(client, structures.OpcodeError, structures.ErrorPayload{
		Message: message,
	})
}

// sendErrorAndClose sends an error and closes the connection
func (m *Manager) sendErrorAndClose(c *websocket.Conn, message string) error {
	errMsg := structures.NewMessage(structures.OpcodeError, structures.ErrorPayload{
		Message: message,
	})
	payload, err := structures.MarshalMessage(errMsg)
	if err == nil {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}
	return c.Close()
}

// BroadcastToAll sends a message to all connected clients
func (m *Manager) BroadcastToAll(op structures.Opcode, data interface{}) {
	msg := structures.NewMessage(op, data)
	payload, err := structures.MarshalMessage(msg)
	if err != nil {
		slog.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	m.clientsMutex.RLock()
	connectedCount := len(m.clients)
	clientList := make([]string, 0, len(m.clients))
	for userID := range m.clients {
		clientList = append(clientList, userID)
	}
	m.clientsMutex.RUnlock()

	slog.Info("📡 Broadcasting WebSocket message",
		"opcode", op,
		"opcodeType", op.String(),
		"connectedClients", connectedCount,
		"clientList", clientList,
		"messageSize", len(payload))

	if connectedCount == 0 {
		slog.Warn("⚠️ No connected clients to broadcast to")
		return
	}

	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	var failedClients []*Client
	sentCount := 0
	for userID, client := range m.clients {
		select {
		case client.sendChan <- payload:
			// Message sent successfully
			sentCount++
			slog.Debug("✅ Message sent to client", "userID", userID)
		default:
			failedClients = append(failedClients, client)
			slog.Warn("❌ Failed to send message to client", "userID", userID, "reason", "channel full")
		}
	}

	slog.Info("📡 WebSocket broadcast complete",
		"sent", sentCount,
		"failed", len(failedClients),
		"opcode", op.String())

	// Clean up failed clients
	for _, client := range failedClients {
		slog.Warn("Removing failed client", "userID", client.UserID)
		go m.unregisterClient(client)
	}
}

// SendToUser sends a message to a specific user
func (m *Manager) SendToUser(userID string, op structures.Opcode, data interface{}) error {
	m.clientsMutex.RLock()
	client, exists := m.clients[userID]
	m.clientsMutex.RUnlock()

	if !exists {
		return &ClientError{Message: "User not connected"}
	}

	return m.sendMessage(client, op, data)
}

// GetConnectedUsers returns a list of connected user IDs
func (m *Manager) GetConnectedUsers() []string {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	users := make([]string, 0, len(m.clients))
	for userID := range m.clients {
		users = append(users, userID)
	}
	return users
}

// getConnectionCount returns the current number of connections
func (m *Manager) getConnectionCount() int {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return len(m.clients)
}

// closeClient closes a client connection
func (m *Manager) closeClient(client *Client) {
	client.cancel()
	client.Conn.Close()
	client.closeOnce.Do(func() {
		close(client.sendChan)
	})
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown() {
	slog.Info("Shutting down WebSocket manager...")

	m.cancel()

	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	for _, client := range m.clients {
		m.closeClient(client)
	}

	m.clients = make(map[string]*Client)
	m.connections = make(map[*websocket.Conn]*Client)

	slog.Info("WebSocket manager shutdown complete")
}

// Custom error types
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}

// Legacy functions for backward compatibility
var defaultManager *Manager

func RegisterRoutes(gctx global.Context, router fiber.Router) {
	if defaultManager == nil {
		slog.Info("🔧 Creating new WebSocket manager")
		defaultManager = NewManager(gctx.Crate().AuthService)
	}
	slog.Info("🔗 Registering WebSocket routes with defaultManager")
	defaultManager.RegisterRoutes(gctx, router)
}

func BroadcastToAll(op structures.Opcode, data interface{}) {
	if defaultManager != nil {
		slog.Debug("🔄 Calling defaultManager.BroadcastToAll", "opcode", op.String())
		defaultManager.BroadcastToAll(op, data)
	} else {
		slog.Warn("Cannot broadcast WebSocket message: defaultManager is nil (WebSocket not initialized)")
	}
}

func SendToUser(userID string, op structures.Opcode, data interface{}) {
	if defaultManager != nil {
		_ = defaultManager.SendToUser(userID, op, data)
	}
}

func CloseAllConnections() {
	if defaultManager != nil {
		defaultManager.Shutdown()
	}
}

// GetConnectionCount returns the current number of connections
func GetConnectionCount() int {
	if defaultManager != nil {
		return defaultManager.getConnectionCount()
	}
	return 0
}

// GetConnectedUsers returns a list of connected user IDs
func GetConnectedUsers() []string {
	if defaultManager != nil {
		return defaultManager.GetConnectedUsers()
	}
	return []string{}
}

// Helper functions for common message types

// BroadcastDownloadProgress broadcasts download progress to all clients
func BroadcastDownloadProgress(download structures.DownloadProgressPayload) {
	BroadcastToAll(structures.OpcodeDownloadProgress, download)
}

// BroadcastDownloadProgressBatch broadcasts batch download progress to all clients
func BroadcastDownloadProgressBatch(downloads []structures.DownloadProgressPayload) {
	if defaultManager == nil {
		slog.Warn("Cannot broadcast download progress batch: WebSocket manager not initialized")
		return
	}

	payload := structures.DownloadProgressBatchPayload{
		Downloads: downloads,
		Count:     len(downloads),
		Timestamp: time.Now().UnixMilli(),
	}

	slog.Debug("Broadcasting download progress batch",
		"downloadCount", len(downloads),
		"payloadSize", len(payload.Downloads),
		"connectedClients", GetConnectionCount())

	BroadcastToAll(structures.OpcodeDownloadProgressBatch, payload)
}

// BroadcastSystemStatus broadcasts system status to all clients
func BroadcastSystemStatus(status structures.SystemStatusPayload) {
	BroadcastToAll(structures.OpcodeSystemStatus, status)
}

// BroadcastUserActivity broadcasts user activity to all clients
func BroadcastUserActivity(activity structures.UserActivityPayload) {
	BroadcastToAll(structures.OpcodeUserActivity, activity)
}
