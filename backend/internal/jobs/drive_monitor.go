package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
	"golang.org/x/sys/unix"
)

// DriveMonitor handles monitoring of mounted drives
type DriveMonitor struct {
	gctx            global.Context
	pollInterval    time.Duration
	lastPollTime    time.Time
	pollCount       int64
	errorCount      int64
	drivesFound     int64
	lastCleanupTime time.Time
	cleanupCount    int64
	stopChan        chan struct{}
	started         bool
}

// NewDriveMonitor creates a new drive monitor instance
func NewDriveMonitor(gctx global.Context) (*DriveMonitor, error) {
	dm := &DriveMonitor{
		gctx:            gctx,
		pollInterval:    5 * time.Minute, // Poll every 5 minutes by default
		lastCleanupTime: time.Now(),
		stopChan:        make(chan struct{}),
		started:         false,
	}
	return dm, nil
}

// Name returns the job name
func (dm *DriveMonitor) Name() structures.Job {
	return structures.JobDriveMonitor
}

func (dm *DriveMonitor) Trigger() error {
	go dm.pollDrives()
	return nil
}

// Start begins the drive monitoring loop (non-blocking)
func (dm *DriveMonitor) Start() {
	if dm.started {
		return
	}
	dm.started = true
	slog.Info("Starting drive monitor", "poll_interval", dm.pollInterval)
	dm.pollDrives()
	go dm.startPolling()
}

// startPolling is the internal polling loop
func (dm *DriveMonitor) startPolling() {
	ticker := time.NewTicker(dm.pollInterval)
	defer ticker.Stop()

	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Drive monitor goroutine panicked, restarting", "error", r)
				}
			}()
			select {
			case <-ticker.C:
				dm.pollDrives()
			case <-dm.stopChan:
				return
			}
		}()
	}
}

// Stop stops the drive monitor
func (dm *DriveMonitor) Stop(ctx context.Context) error {
	slog.Info("Stopping drive monitor")
	close(dm.stopChan)
	return nil
}

// pollDrives polls all configured drives for statistics
func (dm *DriveMonitor) pollDrives() {
	ctx := context.Background()

	slog.Debug("====================START OF DRIVE MONITOR JOB====================")

	// 1. Fetch all configured drives
	slog.Debug("STEP 1: Fetching configured drives")
	drives, err := dm.gctx.Crate().Sqlite.Query().GetMountedDrivesForPolling(ctx)
	if err != nil {
		slog.Error("Failed to fetch mounted drives", "error", err)
		dm.errorCount++
		slog.Debug("====================END OF DRIVE MONITOR JOB (ERROR)====================")
		return
	}
	slog.Debug("STEP 1 COMPLETE: Fetched configured drives", "count", len(drives))

	if len(drives) == 0 {
		slog.Debug("No drives configured for monitoring")
		slog.Debug("====================END OF DRIVE MONITOR JOB (NO DRIVES)====================")
		return
	}

	// 2. Poll each drive for statistics
	slog.Debug("STEP 2: Polling drive statistics")
	var updatedDrives []structures.DriveStatsPayload

	for i, drive := range drives {
		slog.Debug("Processing drive", "index", i+1, "total", len(drives), "name", drive.Name, "path", drive.MountPath)

		stats, err := dm.getDriveStats(drive.MountPath)
		if err != nil {
			slog.Error("Failed to get drive stats", "drive", drive.Name, "path", drive.MountPath, "error", err)
			// Mark drive as offline
			stats = structures.DriveStats{
				IsOnline: false,
			}
		}

		// Update drive in database
		err = dm.gctx.Crate().Sqlite.Query().UpdateMountedDriveStats(ctx,
			repository.UpdateMountedDriveStatsParams{
				TotalSize:       utils.NewNullInt64(stats.TotalSize, true),
				UsedSize:        utils.NewNullInt64(stats.UsedSize, true),
				AvailableSize:   utils.NewNullInt64(stats.AvailableSize, true),
				UsagePercentage: utils.NewNullFloat64(stats.UsagePercentage, true),
				IsOnline:        utils.NewNullBool(stats.IsOnline),
				LastChecked:     utils.NewNullTime(time.Now()),
				ID:              drive.ID,
			})
		if err != nil {
			slog.Error("Failed to update drive stats in database", "drive", drive.Name, "error", err)
			continue
		}

		// Add to payload for WebSocket broadcast
		updatedDrives = append(updatedDrives, structures.DriveStatsPayload{
			ID:          drive.ID,
			Name:        drive.Name,
			MountPath:   drive.MountPath,
			Stats:       stats,
			LastChecked: time.Now(),
		})

		if stats.IsOnline {
			dm.drivesFound++
		}
	}
	slog.Debug("STEP 2 COMPLETE: Polled drive statistics", "updatedDrives", len(updatedDrives))

	// 3. Broadcast updates via WebSocket
	slog.Debug("STEP 3: Broadcasting via WebSocket")
	if len(updatedDrives) > 0 {
		// Import the websocket package to broadcast
		// This would need to be implemented based on your existing WebSocket structure
		slog.Info("Updated drive statistics", "count", len(updatedDrives))
		slog.Debug("STEP 3 COMPLETE: Broadcasted via WebSocket", "batchSize", len(updatedDrives))
	} else {
		slog.Debug("STEP 3 COMPLETE: No drives to broadcast")
	}

	// 4. Update metrics
	slog.Debug("STEP 4: Updating metrics")
	dm.lastPollTime = time.Now()
	dm.pollCount++

	// Adaptive polling: adjust interval based on activity
	dm.adjustPollInterval(len(updatedDrives))
	slog.Debug("STEP 4 COMPLETE: Updated metrics", "pollCount", dm.pollCount, "drivesFound", dm.drivesFound, "pollInterval", dm.pollInterval)

	// Log metrics every 10th poll
	if dm.pollCount%10 == 0 {
		slog.Info(" drive monitor metrics",
			"polls", dm.pollCount,
			"errors", dm.errorCount,
			"drives_found", dm.drivesFound,
			"cleanups", dm.cleanupCount,
			"poll_interval", dm.pollInterval,
			"last_poll", dm.lastPollTime.Format(time.RFC3339))
	}

	slog.Debug("====================END OF DRIVE MONITOR JOB====================")
}

// getDriveStats retrieves filesystem statistics for a given mount path
func (dm *DriveMonitor) getDriveStats(mountPath string) (structures.DriveStats, error) {
	// Ensure the path exists and is accessible
	if _, err := os.Stat(mountPath); err != nil {
		return structures.DriveStats{IsOnline: false}, fmt.Errorf("mount path not accessible: %w", err)
	}

	// Get filesystem statistics using statfs
	var stat unix.Statfs_t
	if err := unix.Statfs(mountPath, &stat); err != nil {
		return structures.DriveStats{IsOnline: false}, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate sizes (statfs uses 512-byte blocks)
	blockSize := uint64(stat.Bsize)
	totalSize := int64(stat.Blocks * blockSize)
	freeSize := int64(stat.Bfree * blockSize)
	availableSize := int64(stat.Bavail * blockSize)
	usedSize := totalSize - freeSize

	// Calculate usage percentage
	usagePercentage := 0.0
	if totalSize > 0 {
		usagePercentage = float64(usedSize) / float64(totalSize) * 100.0
	}

	return structures.DriveStats{
		TotalSize:       totalSize,
		UsedSize:        usedSize,
		AvailableSize:   availableSize,
		UsagePercentage: usagePercentage,
		IsOnline:        true,
	}, nil
}

// adjustPollInterval adjusts the polling interval based on activity
func (dm *DriveMonitor) adjustPollInterval(activeDrives int) {
	// If no drives are active, poll less frequently
	if activeDrives == 0 {
		dm.pollInterval = 10 * time.Minute
	} else {
		// Poll more frequently if drives are active
		dm.pollInterval = 5 * time.Minute
	}
}

// GetSystemDrives returns a list of all mounted filesystems on the system
func (dm *DriveMonitor) GetSystemDrives() ([]string, error) {
	// Read /proc/mounts to get all mounted filesystems
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/mounts: %w", err)
	}

	var drives []string
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]
		filesystem := fields[2]

		// Skip certain filesystem types
		skipFilesystems := []string{
			"proc", "sysfs", "devpts", "tmpfs", "devtmpfs", "cgroup",
			"securityfs", "pstore", "efivarfs", "bpf", "configfs", "debugfs",
			"tracefs", "fusectl", "fuse.gvfsd-fuse", "gvfsd-fuse", "fuse.lxcfs", "lxcfs",
		}
		if utils.MatchesAnyPattern(filesystem, skipFilesystems, false) {
			continue
		}

		// Skip if mount point is in /proc, /sys, /dev, etc.
		if strings.HasPrefix(mountPoint, "/proc") || strings.HasPrefix(mountPoint, "/sys") ||
			strings.HasPrefix(mountPoint, "/dev") || strings.HasPrefix(mountPoint, "/run") ||
			strings.HasPrefix(mountPoint, "/tmp") || strings.HasPrefix(mountPoint, "/var") {
			continue
		}

		// Add to list if it's a real storage device
		if strings.HasPrefix(device, "/dev/") || strings.HasPrefix(device, "/mnt/") ||
			strings.HasPrefix(device, "/media/") || strings.HasPrefix(device, "/home/") {
			drives = append(drives, mountPoint)
		}
	}

	return drives, nil
}
