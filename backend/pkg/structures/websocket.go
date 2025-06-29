package structures

import (
	"encoding/json"
	"fmt"
	"time"
)

// --- OPCODES ---

type Opcode uint8

const (
	OpcodeDispatch  Opcode = 0 // Server sends event
	OpcodeHello     Opcode = 1 // Server greets client
	OpcodeHeartbeat Opcode = 2 // Server/client keepalive
	OpcodeReconnect Opcode = 3 // Server requests reconnect
	OpcodeAck       Opcode = 4 // Server acknowledges action
	OpcodeError     Opcode = 5 // Server sends error

	OpcodeDownloadProgress      Opcode = 10 // Server sends download progress
	OpcodeDownloadRemoved       Opcode = 11 // Server notifies download removal
	OpcodeDownloadProgressBatch Opcode = 12 // Server sends batch download progress
	OpcodeSystemStatus          Opcode = 13 // Server sends system status
	OpcodeUserActivity          Opcode = 14 // Server sends user activity updates
)

// String returns the string representation of an opcode
func (o Opcode) String() string {
	switch o {
	case OpcodeDispatch:
		return "Dispatch"
	case OpcodeHello:
		return "Hello"
	case OpcodeHeartbeat:
		return "Heartbeat"
	case OpcodeReconnect:
		return "Reconnect"
	case OpcodeAck:
		return "Ack"
	case OpcodeError:
		return "Error"
	case OpcodeDownloadProgress:
		return "DownloadProgress"
	case OpcodeDownloadRemoved:
		return "DownloadRemoved"
	case OpcodeDownloadProgressBatch:
		return "DownloadProgressBatch"
	case OpcodeSystemStatus:
		return "SystemStatus"
	case OpcodeUserActivity:
		return "UserActivity"
	default:
		return fmt.Sprintf("Unknown(%d)", o)
	}
}

// IsValid checks if the opcode is valid
func (o Opcode) IsValid() bool {
	return o <= OpcodeError || (o >= OpcodeDownloadProgress && o <= OpcodeUserActivity)
}

// --- WRAPPED MESSAGE ---

type Message struct {
	Op        Opcode      `json:"op"`          // Operation type
	Timestamp int64       `json:"t"`           // Millisecond timestamp
	Data      interface{} `json:"d"`           // Any payload
	Sequence  *uint64     `json:"s,omitempty"` // Optional sequence number
}

// NewMessage creates a new message with validation
func NewMessage(op Opcode, data interface{}) Message {
	if !op.IsValid() {
		panic(fmt.Sprintf("invalid opcode: %d", op))
	}

	return Message{
		Op:        op,
		Timestamp: time.Now().UnixMilli(),
		Data:      data,
	}
}

// NewMessageWithSequence creates a new message with a sequence number
func NewMessageWithSequence(op Opcode, data interface{}, sequence uint64) Message {
	msg := NewMessage(op, data)
	msg.Sequence = &sequence
	return msg
}

// Validate checks if the message is valid
func (m Message) Validate() error {
	if !m.Op.IsValid() {
		return fmt.Errorf("invalid opcode: %d", m.Op)
	}

	if m.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp: %d", m.Timestamp)
	}

	// Check if timestamp is not too far in the future (5 minutes)
	if m.Timestamp > time.Now().Add(5*time.Minute).UnixMilli() {
		return fmt.Errorf("timestamp too far in future: %d", m.Timestamp)
	}

	return nil
}

// --- PAYLOAD TYPES ---

// HelloPayload is sent when a client connects
type HelloPayload struct {
	Message  string            `json:"message"`
	ServerID string            `json:"server_id,omitempty"`
	Features []string          `json:"features,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// ErrorPayload represents an error message
type ErrorPayload struct {
	Message   string `json:"message"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// DownloadProgressPayload represents download progress
type DownloadProgressPayload struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	TorrentTitle string  `json:"torrent_title"`
	Source       string  `json:"source"`
	TMDBID       *int64  `json:"tmdb_id,omitempty"` // Optional TMDB ID
	TvDBID       *int64  `json:"tvdb_id,omitempty"` // Optional TVDB ID
	Hash         string  `json:"hash"`
	Progress     float64 `json:"progress"` // 0-100
	TimeLeft     string  `json:"time_left"`
	Status       string  `json:"status"`
	LastUpdated  string  `json:"last_updated"`
}

// DownloadRemovedPayload represents a removed download
type DownloadRemovedPayload struct {
	DownloadID string `json:"download_id"`
	Reason     string `json:"reason,omitempty"`
}

// DownloadProgressBatchPayload represents multiple download updates
type DownloadProgressBatchPayload struct {
	Downloads []DownloadProgressPayload `json:"downloads"`
	Count     int                       `json:"count"`
	Timestamp int64                     `json:"timestamp"`
}

// SystemStatusPayload represents system status information
type SystemStatusPayload struct {
	Status      string            `json:"status"`           // "online", "maintenance", "error"
	Uptime      int64             `json:"uptime"`           // Server uptime in seconds
	Connections int               `json:"connections"`      // Active WebSocket connections
	Load        map[string]string `json:"load,omitempty"`   // System load information
	Memory      map[string]int64  `json:"memory,omitempty"` // Memory usage
	Disk        map[string]int64  `json:"disk,omitempty"`   // Disk usage
}

// UserActivityPayload represents user activity updates
type UserActivityPayload struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Activity  string                 `json:"activity"` // "login", "logout", "download_start", etc.
	Timestamp int64                  `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// --- VALIDATION HELPERS ---

// ValidateDownloadProgress validates download progress data
func (d DownloadProgressPayload) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("id is required")
	}

	if d.Progress < 0 || d.Progress > 100 {
		return fmt.Errorf("progress must be between 0 and 100, got %f", d.Progress)
	}

	return nil
}

// ValidateSystemStatus validates system status data
func (s SystemStatusPayload) Validate() error {
	validStatuses := map[string]bool{
		"online":      true,
		"maintenance": true,
		"error":       true,
	}

	if !validStatuses[s.Status] {
		return fmt.Errorf("invalid status: %s", s.Status)
	}

	if s.Uptime < 0 {
		return fmt.Errorf("uptime cannot be negative")
	}

	if s.Connections < 0 {
		return fmt.Errorf("connections cannot be negative")
	}

	return nil
}

// --- UTILITY FUNCTIONS ---

// ParseMessage parses a JSON message and validates it
func ParseMessage(data []byte) (Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	if err := msg.Validate(); err != nil {
		return Message{}, fmt.Errorf("invalid message: %w", err)
	}

	return msg, nil
}

// MarshalMessage marshals a message to JSON with validation
func MarshalMessage(msg Message) ([]byte, error) {
	if err := msg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	return json.Marshal(msg)
}
