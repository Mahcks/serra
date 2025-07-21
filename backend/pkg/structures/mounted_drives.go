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

// UpdateDriveThresholdsRequest represents a request to update drive monitoring thresholds
type UpdateDriveThresholdsRequest struct {
	MonitoringEnabled   *bool    `json:"monitoring_enabled,omitempty"`
	WarningThreshold    *float64 `json:"warning_threshold,omitempty"`
	CriticalThreshold   *float64 `json:"critical_threshold,omitempty"`
	GrowthRateThreshold *float64 `json:"growth_rate_threshold,omitempty"`
}

// DriveThresholds represents the current threshold configuration for a drive
type DriveThresholds struct {
	DriveID             string   `json:"drive_id"`
	MonitoringEnabled   bool     `json:"monitoring_enabled"`
	WarningThreshold    float64  `json:"warning_threshold"`
	CriticalThreshold   float64  `json:"critical_threshold"`
	GrowthRateThreshold float64  `json:"growth_rate_threshold"`
}

// StoragePool represents a storage pool (ZFS, UnRAID, etc.)
type StoragePool struct {
	Name            string                 `json:"name"`
	Type            string                 `json:"type"` // "zfs", "unraid", "raid", "lvm"
	Health          string                 `json:"health"`
	Status          string                 `json:"status"`
	TotalSize       int64                  `json:"total_size"`
	UsedSize        int64                  `json:"used_size"`
	AvailableSize   int64                  `json:"available_size"`
	UsagePercentage float64                `json:"usage_percentage"`
	Redundancy      string                 `json:"redundancy,omitempty"`
	Devices         []StoragePoolDevice    `json:"devices,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
	LastChecked     time.Time              `json:"last_checked"`
}

// StoragePoolDevice represents a device within a storage pool
type StoragePoolDevice struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Status   string `json:"status"`
	Health   string `json:"health,omitempty"`
	Size     int64  `json:"size,omitempty"`
	ReadErrors  int64  `json:"read_errors,omitempty"`
	WriteErrors int64  `json:"write_errors,omitempty"`
	ChecksumErrors int64 `json:"checksum_errors,omitempty"`
}

// ZFSPool represents a ZFS pool with specific ZFS properties
type ZFSPool struct {
	StoragePool
	Compression     string  `json:"compression,omitempty"`
	Deduplication   string  `json:"deduplication,omitempty"`
	ScrubStatus     string  `json:"scrub_status,omitempty"`
	ScrubProgress   float64 `json:"scrub_progress,omitempty"`
	FragmentationPct float64 `json:"fragmentation_pct,omitempty"`
}

// UnRAIDArray represents an UnRAID array
type UnRAIDArray struct {
	StoragePool
	ParityDevices []StoragePoolDevice `json:"parity_devices,omitempty"`
	DataDevices   []StoragePoolDevice `json:"data_devices,omitempty"`
	CacheDevices  []StoragePoolDevice `json:"cache_devices,omitempty"`
	SyncStatus    string              `json:"sync_status,omitempty"`
	SyncProgress  float64             `json:"sync_progress,omitempty"`
}
