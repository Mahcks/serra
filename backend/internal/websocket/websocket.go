package websocket

import (
	"context"
	"encoding/json"
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

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Services
	authService auth.Authmen
}

// Client represents a connected WebSocket client
type Client struct {
	Conn        *websocket.Conn
	User        *auth.JWTClaimUser
	UserID      string
	ConnectedAt time.Time
	LastPing    time.Time

	// Channel for sending messages
	sendChan chan []byte

	// Context for this client's lifecycle
	ctx    context.Context
	cancel context.CancelFunc

	// Mutex for thread-safe operations
	mu        sync.Mutex
	closeOnce sync.Once
}

// NewManager creates a new WebSocket manager
func NewManager(authService auth.Authmen) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		clients:           make(map[string]*Client),
		connections:       make(map[*websocket.Conn]*Client),
		maxConnections:    1000, // Configurable
		heartbeatInterval: 30 * time.Second,
		connectionTimeout: 2 * time.Minute,
		ctx:               ctx,
		cancel:            cancel,
		authService:       authService,
	}
}

// RegisterRoutes sets up the websocket endpoint
func (m *Manager) RegisterRoutes(gctx global.Context, router fiber.Router) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
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
		slog.Warn("Connection limit reached")
		m.sendErrorAndClose(c, "Server at capacity")
		return
	}

	// Extract and validate token
	client, err := m.authenticateConnection(c)
	if err != nil {
		slog.Warn("Authentication failed", "error", err)
		m.sendErrorAndClose(c, err.Error())
		return
	}

	// Register client
	if !m.registerClient(client) {
		slog.Warn("Failed to register client", "userID", client.UserID)
		m.sendErrorAndClose(c, "Failed to register connection")
		return
	}
	// defer m.unregisterClient(client)

	// Send welcome message
	m.sendMessage(client, structures.OpcodeAck, structures.HelloPayload{
		Message: "Connected successfully",
	})

	slog.Info("WebSocket connected", "userID", client.UserID, "username", client.User.Username)

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
		return nil, &AuthError{Message: "Invalid auth token"}
	}

	// Create client context
	ctx, cancel := context.WithCancel(m.ctx)

	client := &Client{
		Conn:        c,
		User:        claims,
		UserID:      claims.UserID,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		sendChan:    make(chan []byte, 100), // Buffer for 100 messages
		ctx:         ctx,
		cancel:      cancel,
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

	slog.Info("WebSocket disconnected", "userID", client.UserID)
}

// handleClientMessages processes incoming messages from a client
func (m *Manager) handleClientMessages(client *Client) {
	defer m.unregisterClient(client)

	// Set read deadline
	client.Conn.SetReadDeadline(time.Now().Add(m.connectionTimeout))

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
			client.Conn.SetReadDeadline(time.Now().Add(m.connectionTimeout))
		}
	}
}

// handleMessage processes a single message from a client
func (m *Manager) handleMessage(client *Client, message []byte) error {
	var msg structures.Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return m.sendError(client, "Invalid message format")
	}

	switch msg.Op {
	case structures.OpcodeHeartbeat:
		client.mu.Lock()
		client.LastPing = time.Now()
		client.mu.Unlock()

		// Send pong response
		return m.sendMessage(client, structures.OpcodeHeartbeat, nil)

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
			client.mu.Unlock()

			// Check if client is responsive
			if time.Since(lastPing) > m.connectionTimeout {
				slog.Warn("Client timeout", "userID", client.UserID)
				return
			}

			// Send heartbeat
			if err := m.sendMessage(client, structures.OpcodeHeartbeat, nil); err != nil {
				slog.Error("Failed to send heartbeat", "userID", client.UserID, "error", err)
				return
			}
		}
	}
}

// sendMessage sends a message to a specific client
func (m *Manager) sendMessage(client *Client, op structures.Opcode, data interface{}) error {
	msg := structures.NewMessage(op, data)
	payload, err := json.Marshal(msg)
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
	_ = c.WriteJSON(errMsg)
	return c.Close()
}

// BroadcastToAll sends a message to all connected clients
func (m *Manager) BroadcastToAll(op structures.Opcode, data interface{}) {
	msg := structures.NewMessage(op, data)
	payload, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal broadcast message", "error", err)
		return
	}

	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	var failedClients []*Client
	for _, client := range m.clients {
		select {
		case client.sendChan <- payload:
			// Message sent successfully
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
