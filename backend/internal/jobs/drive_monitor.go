package jobs

import (
	"context"
	"database/sql"
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
	*BaseJob
	lastPollTime    time.Time
	drivesFound     int64
	lastCleanupTime time.Time
	cleanupCount    int64
}

// NewDriveMonitor creates a new drive monitor instance
func NewDriveMonitor(gctx global.Context, config JobConfig) (*DriveMonitor, error) {
	base := NewBaseJob(gctx, structures.JobDriveMonitor, config)
	dm := &DriveMonitor{
		BaseJob:         base,
		lastCleanupTime: time.Now(),
	}
	return dm, nil
}

// Trigger executes the drive polling task
func (dm *DriveMonitor) Trigger(ctx context.Context) error {
	return dm.pollDrivesWithContext(ctx)
}

// Start begins the drive monitoring (delegates to BaseJob)
func (dm *DriveMonitor) Start(ctx context.Context) error {
	slog.Info("Starting drive monitor", "interval", dm.Config().Interval)
	return dm.BaseJob.Start(ctx)
}

// pollDrivesWithContext polls all configured drives for statistics with context support
func (dm *DriveMonitor) pollDrivesWithContext(ctx context.Context) error {

	slog.Debug("====================START OF DRIVE MONITOR JOB====================")

	// 1. Fetch all configured drives
	slog.Debug("STEP 1: Fetching configured drives")
	drives, err := dm.Context().Crate().Sqlite.Query().GetMountedDrivesForPolling(ctx)
	if err != nil {
		slog.Error("Failed to fetch mounted drives", "error", err)
		slog.Debug("====================END OF DRIVE MONITOR JOB (ERROR)====================")
		return err
	}
	slog.Debug("STEP 1 COMPLETE: Fetched configured drives", "count", len(drives))

	if len(drives) == 0 {
		slog.Debug("No drives configured for monitoring")
		slog.Debug("====================END OF DRIVE MONITOR JOB (NO DRIVES)====================")
		return nil
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
		err = dm.Context().Crate().Sqlite.Query().UpdateMountedDriveStats(ctx,
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

		// Record usage history for analytics if drive is online
		if stats.IsOnline {
			err = dm.recordDriveUsageAnalytics(ctx, drive, stats)
			if err != nil {
				slog.Error("Failed to record drive analytics", "drive", drive.Name, "error", err)
				// Continue processing other drives even if analytics fails
			}
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

	slog.Debug("STEP 4 COMPLETE: Updated metrics", "drivesFound", dm.drivesFound, "interval", dm.Config().Interval)

	// Log metrics periodically
	metrics := dm.Metrics()
	if metrics.RunCount%10 == 0 {
		slog.Info("Drive monitor metrics",
			"runs", metrics.RunCount,
			"errors", metrics.ErrorCount,
			"drives_found", dm.drivesFound,
			"cleanups", dm.cleanupCount,
			"interval", dm.Config().Interval,
			"last_run", metrics.LastRun.Format(time.RFC3339))
	}

	slog.Debug("====================END OF DRIVE MONITOR JOB====================")
	return nil
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

// recordDriveUsageAnalytics records drive usage data for analytics and checks for alerts
func (dm *DriveMonitor) recordDriveUsageAnalytics(ctx context.Context, drive repository.MountedDrife, stats structures.DriveStats) error {
	// Calculate growth rate from historical data
	growthRate, projectedFullDate, err := dm.calculateGrowthRate(ctx, drive.ID, stats.UsedSize)
	if err != nil {
		slog.Warn("Failed to calculate growth rate", "drive", drive.Name, "error", err)
		// Continue without growth rate data
	}

	// Record the current usage data
	_, err = dm.Context().Crate().Sqlite.Query().RecordDriveUsage(ctx, repository.RecordDriveUsageParams{
		DriveID:            drive.ID,
		TotalSize:          stats.TotalSize,
		UsedSize:           stats.UsedSize,
		AvailableSize:      stats.AvailableSize,
		UsagePercentage:    stats.UsagePercentage,
		GrowthRateGbPerDay: sql.NullFloat64{Float64: growthRate, Valid: growthRate > 0},
		ProjectedFullDate:  sql.NullTime{Time: timeFromDateString(projectedFullDate), Valid: projectedFullDate != nil},
	})
	if err != nil {
		return fmt.Errorf("failed to record drive usage: %w", err)
	}

	// Check for alerts based on usage thresholds
	return dm.checkDriveAlerts(ctx, drive, stats, growthRate, projectedFullDate)
}

// calculateGrowthRate calculates the daily growth rate based on historical data
func (dm *DriveMonitor) calculateGrowthRate(ctx context.Context, driveID string, currentUsedSize int64) (float64, *string, error) {
	// Get usage history for analysis (limit to recent records for performance)
	history, err := dm.Context().Crate().Sqlite.Query().GetDriveUsageHistory(ctx, repository.GetDriveUsageHistoryParams{
		DriveID: driveID,
		Limit:   50, // Get last 50 records for trend analysis
	})
	if err != nil {
		return 0, nil, err
	}

	if len(history) < 2 {
		// Not enough data to calculate growth rate
		return 0, nil, nil
	}

	// Calculate average daily growth rate from the last week's data
	var totalGrowthGB float64
	var validDataPoints int

	for i := 1; i < len(history) && i < 10; i++ { // Use up to 10 recent data points
		prev := history[i]   // Older entry (DESC order)
		curr := history[i-1] // Newer entry

		if !curr.RecordedAt.Valid || !prev.RecordedAt.Valid {
			continue
		}

		timeDiff := curr.RecordedAt.Time.Sub(prev.RecordedAt.Time)
		if timeDiff.Hours() < 1 { // Skip if too close in time
			continue
		}

		sizeDiffBytes := curr.UsedSize - prev.UsedSize
		if sizeDiffBytes <= 0 { // Only consider positive growth
			continue
		}

		sizeDiffGB := float64(sizeDiffBytes) / (1024 * 1024 * 1024) // Convert to GB
		dailyGrowth := sizeDiffGB / (timeDiff.Hours() / 24)         // GB per day

		totalGrowthGB += dailyGrowth
		validDataPoints++
	}

	if validDataPoints == 0 {
		return 0, nil, nil
	}

	avgGrowthRateGBPerDay := totalGrowthGB / float64(validDataPoints)

	// Calculate projected full date if growth rate is meaningful
	var projectedFullDate *string
	if avgGrowthRateGBPerDay > 0.1 { // Only if growing more than 0.1GB per day
		availableGB := float64(history[0].AvailableSize) / (1024 * 1024 * 1024)
		daysUntilFull := availableGB / avgGrowthRateGBPerDay

		if daysUntilFull > 0 && daysUntilFull < 365*2 { // Only if less than 2 years
			fullDate := time.Now().AddDate(0, 0, int(daysUntilFull))
			projectedDateStr := fullDate.Format("2006-01-02")
			projectedFullDate = &projectedDateStr
		}
	}

	return avgGrowthRateGBPerDay, projectedFullDate, nil
}

// checkDriveAlerts checks if any alerts should be created for the drive
func (dm *DriveMonitor) checkDriveAlerts(ctx context.Context, drive repository.MountedDrife, stats structures.DriveStats, growthRate float64, projectedFullDate *string) error {
	// Skip monitoring if disabled for this drive
	if drive.MonitoringEnabled.Valid && !drive.MonitoringEnabled.Bool {
		slog.Debug("Drive monitoring disabled", "drive", drive.Name, "drive_id", drive.ID)
		return nil
	}

	// Use custom thresholds if available, otherwise fall back to defaults
	warningThreshold := 80.0  // Default warning threshold
	criticalThreshold := 95.0 // Default critical threshold
	growthRateThreshold := 50.0 // Default growth rate threshold

	if drive.WarningThreshold.Valid {
		warningThreshold = drive.WarningThreshold.Float64
	}
	if drive.CriticalThreshold.Valid {
		criticalThreshold = drive.CriticalThreshold.Float64
	}
	if drive.GrowthRateThreshold.Valid {
		growthRateThreshold = drive.GrowthRateThreshold.Float64
	}

	// Check for critical usage threshold
	if stats.UsagePercentage >= criticalThreshold {
		err := dm.createDriveAlert(ctx, drive, "usage_threshold", criticalThreshold, stats.UsagePercentage,
			fmt.Sprintf("CRITICAL: Drive '%s' is %.1f%% full. Immediate action required to free up space.", drive.Name, stats.UsagePercentage))
		if err != nil {
			return err
		}
	} else if stats.UsagePercentage >= warningThreshold {
		// Check for warning usage threshold
		err := dm.createDriveAlert(ctx, drive, "usage_threshold", warningThreshold, stats.UsagePercentage,
			fmt.Sprintf("WARNING: Drive '%s' is %.1f%% full. Consider freeing up space soon.", drive.Name, stats.UsagePercentage))
		if err != nil {
			return err
		}
	}

	// Check growth rate alerts
	if growthRate >= growthRateThreshold {
		err := dm.createDriveAlert(ctx, drive, "growth_rate", growthRateThreshold, growthRate,
			fmt.Sprintf("Drive '%s' is growing rapidly at %.1f GB per day. Monitor closely.", drive.Name, growthRate))
		if err != nil {
			return err
		}
	}

	// Check projected full date alerts
	if projectedFullDate != nil {
		projectedDate, err := time.Parse("2006-01-02", *projectedFullDate)
		if err == nil {
			daysUntilFull := time.Until(projectedDate).Hours() / 24
			if daysUntilFull <= 30 && daysUntilFull > 0 { // Alert if full within 30 days
				err := dm.createDriveAlert(ctx, drive, "projected_full", 30, daysUntilFull,
					fmt.Sprintf("Drive '%s' is projected to be full by %s (in %.0f days).", drive.Name, *projectedFullDate, daysUntilFull))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createDriveAlert creates a new alert if one doesn't already exist for this condition
func (dm *DriveMonitor) createDriveAlert(ctx context.Context, drive repository.MountedDrife, alertType string, threshold, currentValue float64, message string) error {
	// Check if an active alert already exists for this drive and alert type
	existingAlerts, err := dm.Context().Crate().Sqlite.Query().GetActiveDriveAlerts(ctx)
	if err != nil {
		return err
	}

	for _, alert := range existingAlerts {
		if alert.DriveID == drive.ID && alert.AlertType == alertType && alert.IsActive.Bool {
			// Alert already exists, just log it
			slog.Debug("Alert already exists for drive", "drive", drive.Name, "alert_type", alertType, "current_value", currentValue)
			return nil
		}
	}

	// Create new alert
	_, err = dm.Context().Crate().Sqlite.Query().CreateDriveAlert(ctx, repository.CreateDriveAlertParams{
		DriveID:        drive.ID,
		AlertType:      alertType,
		ThresholdValue: threshold,
		CurrentValue:   currentValue,
		AlertMessage:   message,
	})
	if err != nil {
		return err
	}

	slog.Info("Created drive alert", "drive", drive.Name, "alert_type", alertType, "threshold", threshold, "current_value", currentValue, "message", message)
	return nil
}

// timeFromDateString converts a date string pointer to time.Time
func timeFromDateString(dateStr *string) time.Time {
	if dateStr == nil {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}
