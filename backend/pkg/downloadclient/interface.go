package downloadclient

import (
	"context"
	"time"
)

// Interface defines the contract that all download clients must implement
type Interface interface {
	// GetType returns the client type identifier (e.g., "qbittorrent", "sabnzbd", "deluge")
	GetType() string

	// GetName returns a human-readable name for the client
	GetName() string

	// Connect establishes a connection to the download client
	Connect(ctx context.Context) error

	// Disconnect closes the connection to the download client
	Disconnect(ctx context.Context) error

	// GetDownloads retrieves all active downloads from the client
	GetDownloads(ctx context.Context) ([]Item, error)

	// GetDownloadProgress retrieves progress for a specific download by ID
	GetDownloadProgress(ctx context.Context, downloadID string) (*Progress, error)

	// IsConnected returns whether the client is currently connected
	IsConnected() bool

	// GetConnectionInfo returns connection details for debugging
	GetConnectionInfo() ConnectionInfo
}

// Item represents a download item from any client
type Item struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Hash          string    `json:"hash,omitempty"`
	Progress      float64   `json:"progress"` // 0-100
	Status        string    `json:"status"`
	Size          int64     `json:"size"`
	SizeLeft      int64     `json:"size_left"`
	TimeLeft      string    `json:"time_left"`
	DownloadSpeed string    `json:"download_speed,omitempty"`
	UploadSpeed   string    `json:"upload_speed,omitempty"`
	ETA           int64     `json:"eta,omitempty"` // seconds
	Category      string    `json:"category,omitempty"`
	Tags          []string  `json:"tags,omitempty"`
	AddedOn       time.Time `json:"added_on"`
}

// Progress represents progress information for a download
type Progress struct {
	Progress      float64 `json:"progress"` // 0-100
	TimeLeft      string  `json:"time_left"`
	DownloadSpeed string  `json:"download_speed,omitempty"`
	UploadSpeed   string  `json:"upload_speed,omitempty"`
	Status        string  `json:"status"`
}

// ConnectionInfo provides connection details for debugging
type ConnectionInfo struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	UseSSL    bool   `json:"use_ssl"`
	Connected bool   `json:"connected"`
	LastError string `json:"last_error,omitempty"`
}

// Config holds configuration for a download client
type Config struct {
	ID       string
	Type     string
	Name     string
	Host     string
	Port     int
	Username *string
	Password *string
	APIKey   *string
	UseSSL   bool
}

// UnsupportedClientError is returned when trying to create an unsupported client type
type UnsupportedClientError struct {
	ClientType string
}

func (e *UnsupportedClientError) Error() string {
	return "unsupported download client type: " + e.ClientType
}

// DownloadNotFoundError is returned when a download is not found
type DownloadNotFoundError struct {
	DownloadID string
}

func (e *DownloadNotFoundError) Error() string {
	return "download not found: " + e.DownloadID
}
