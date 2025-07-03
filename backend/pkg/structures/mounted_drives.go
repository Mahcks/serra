package structures

import (
	"time"
)

// MountedDrive represents a mounted filesystem drive
type MountedDrive struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	MountPath       string    `json:"mount_path"`
	Filesystem      *string   `json:"filesystem,omitempty"`
	TotalSize       *int64    `json:"total_size,omitempty"`
	UsedSize        *int64    `json:"used_size,omitempty"`
	AvailableSize   *int64    `json:"available_size,omitempty"`
	UsagePercentage *float64  `json:"usage_percentage,omitempty"`
	IsOnline        bool      `json:"is_online"`
	LastChecked     time.Time `json:"last_checked"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateMountedDriveRequest represents a request to create a new mounted drive
type CreateMountedDriveRequest struct {
	Name      string `json:"name" validate:"required"`
	MountPath string `json:"mount_path" validate:"required"`
}

// UpdateMountedDriveRequest represents a request to update a mounted drive
type UpdateMountedDriveRequest struct {
	Name      string `json:"name" validate:"required"`
	MountPath string `json:"mount_path" validate:"required"`
}

// DriveStats represents the current statistics of a drive
type DriveStats struct {
	TotalSize       int64   `json:"total_size"`
	UsedSize        int64   `json:"used_size"`
	AvailableSize   int64   `json:"available_size"`
	UsagePercentage float64 `json:"usage_percentage"`
	IsOnline        bool    `json:"is_online"`
}

// DriveStatsPayload represents the WebSocket payload for drive statistics updates
type DriveStatsPayload struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	MountPath   string     `json:"mount_path"`
	Stats       DriveStats `json:"stats"`
	LastChecked time.Time  `json:"last_checked"`
}
