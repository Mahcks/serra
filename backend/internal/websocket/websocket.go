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
		if websocket.IsWebSocketUpgrade(c) {
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
		m.handleConnection(c, gctx)
	}))
}

// handleConnection manages a new WebSocket connection
func (m *Manager) handleConnection(c *websocket.Conn, gctx global.Context) {
	// Check connection limit
	if m.getConnectionCount() >= m.maxConnections {
		m.sendErrorAndClose(c, "Server at capacity")
		return
	}

	// Extract and validate token
	client, err := m.authenticateConnection(c)
	if err != nil {
		m.sendErrorAndClose(c, err.Error())
		return
	}

	// Register client
	if !m.registerClient(client) {
		m.sendErrorAndClose(c, "Failed to register connection")
		return
	}

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

	// Start client message handling
	m.handleClientMessages(client)
}

// authenticateConnection validates the connection and creates a client
func (m *Manager) authenticateConnection(c *websocket.Conn) (*Client, error) {
	// Extract token from cookie
	token := c.Cookies("serra_token")
	if token == "" {
		return nil, &AuthError{Message: "Missing auth token"}
	}

	// Validate token
	claims, err := m.authService.ValidateJWT(token)
	if err != nil {
		// Provide more specific error messages for better client handling
		if err.Error() == "token is expired" || err.Error() == "Token is expired" {
			return nil, &AuthError{Message: "Auth token expired"}
		}
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
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	// Check if user already has a connection
	if existing, exists := m.clients[client.UserID]; exists {
		slog.Info("Replacing existing connection", "userID", client.UserID)
		m.closeClient(existing)
	}

	m.clients[client.UserID] = client
	m.connections[client.Conn] = client

	// Start background goroutines for this client
	go m.clientWriter(client)
	go m.clientHeartbeat(client)

	return true
}

// unregisterClient removes a client from the manager
func (m *Manager) unregisterClient(client *Client) {
	m.clientsMutex.Lock()
	defer m.clientsMutex.Unlock()

	delete(m.clients, client.UserID)
	delete(m.connections, client.Conn)

	client.cancel()
	client.Conn.Close()
	client.closeOnce.Do(func() {
		close(client.sendChan)
	})
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
	msg, err := structures.ParseMessage(message)
	if err != nil {
		slog.Warn("Invalid message received", "userID", client.UserID, "error", err, "rawMessage", string(message))
		return m.sendError(client, "Invalid message format")
	}

	switch msg.Op {
	case structures.OpcodeHeartbeat:
		// Client sent us a heartbeat response - they're alive
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
				// Send heartbeat ping
				now := time.Now()
				client.mu.Lock()
				client.awaitingPong = true
				client.LastPong = now // Track when we sent this ping
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
	for _, client := range m.clients {
		select {
		case client.sendChan <- payload:
			// Message sent successfully
			sentCount++
		default:
			failedClients = append(failedClients, client)
		}
	}

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
		defaultManager = NewManager(gctx.Crate().AuthService)
	}
	defaultManager.RegisterRoutes(gctx, router)
}

func BroadcastToAll(op structures.Opcode, data interface{}) {
	if defaultManager != nil {
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
		return
	}

	payload := structures.DownloadProgressBatchPayload{
		Downloads: downloads,
		Count:     len(downloads),
		Timestamp: time.Now().UnixMilli(),
	}

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

// BroadcastToUser broadcasts a message to a specific user
func BroadcastToUser(userID string, op structures.Opcode, data interface{}) {
	if defaultManager == nil {
		return
	}
	defaultManager.BroadcastToUser(userID, op, data)
}

// BroadcastToUser broadcasts a message to a specific user
func (m *Manager) BroadcastToUser(userID string, op structures.Opcode, data interface{}) {
	m.clientsMutex.RLock()
	client, exists := m.clients[userID]
	m.clientsMutex.RUnlock()

	if !exists || client == nil {
		slog.Debug("User not connected for broadcast", "user_id", userID, "opcode", op)
		return
	}

	msg := structures.NewMessage(op, data)
	payload, err := structures.MarshalMessage(msg)
	if err != nil {
		slog.Error("Failed to marshal user broadcast message", "user_id", userID, "error", err)
		return
	}

	select {
	case client.sendChan <- payload:
		slog.Debug("Broadcast sent to user", "user_id", userID, "opcode", op)
	default:
		slog.Warn("Failed to send broadcast to user (channel full)", "user_id", userID, "opcode", op)
	}
}
