package drive_monitor

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
)

type DriveMonitorService struct {
	repo *repository.Queries
}

type DriveUsageData struct {
	DriveID            string    `json:"drive_id"`
	Name               string    `json:"name"`
	MountPath          string    `json:"mount_path"`
	TotalSize          int64     `json:"total_size"`
	UsedSize           int64     `json:"used_size"`
	AvailableSize      int64     `json:"available_size"`
	UsagePercentage    float64   `json:"usage_percentage"`
	GrowthRateGBPerDay float64   `json:"growth_rate_gb_per_day"`
	ProjectedFullDate  *string   `json:"projected_full_date,omitempty"`
	RecordedAt         time.Time `json:"recorded_at"`
}

type DriveAlert struct {
	ID              int64   `json:"id"`
	DriveID         string  `json:"drive_id"`
	DriveName       string  `json:"drive_name"`
	MountPath       string  `json:"mount_path"`
	AlertType       string  `json:"alert_type"`
	ThresholdValue  float64 `json:"threshold_value"`
	CurrentValue    float64 `json:"current_value"`
	AlertMessage    string  `json:"alert_message"`
	IsActive        bool    `json:"is_active"`
	LastTriggered   string  `json:"last_triggered"`
	AcknowledgeCount int64  `json:"acknowledge_count"`
}

// Default alert thresholds
const (
	DefaultUsageThreshold     = 80.0  // 80% usage
	CriticalUsageThreshold    = 95.0  // 95% usage
	DefaultGrowthRateThreshold = 50.0 // 50GB per day growth rate
)

func NewDriveMonitorService(repo *repository.Queries) *DriveMonitorService {
	return &DriveMonitorService{
		repo: repo,
	}
}

// MonitorAllDrives checks usage for all mounted drives and creates alerts if needed
func (s *DriveMonitorService) MonitorAllDrives(ctx context.Context) error {
	slog.Info("Starting drive monitoring cycle")

	// Get all mounted drives
	drives, err := s.repo.ListMountedDrives(ctx)
	if err != nil {
		slog.Error("Failed to get mounted drives", "error", err)
		return fmt.Errorf("failed to get mounted drives: %w", err)
	}

	var processedCount int
	var errorCount int

	for _, drive := range drives {
		if err := s.monitorSingleDrive(ctx, drive); err != nil {
			slog.Error("Failed to monitor drive",
				"drive_id", drive.ID,
				"drive_name", drive.Name,
				"mount_path", drive.MountPath,
				"error", err)
			errorCount++
			continue
		}
		processedCount++
	}

	slog.Info("Drive monitoring cycle completed",
		"processed", processedCount,
		"errors", errorCount,
		"total_drives", len(drives))

	return nil
}

// monitorSingleDrive monitors a specific drive and records usage data
func (s *DriveMonitorService) monitorSingleDrive(ctx context.Context, drive repository.MountedDrife) error {
	// Get current usage from filesystem
	usageData, err := s.getDriveUsage(drive.MountPath)
	if err != nil {
		slog.Error("Failed to get drive usage",
			"drive_id", drive.ID,
			"mount_path", drive.MountPath,
			"error", err)
		return err
	}

	// Calculate growth rate based on historical data
	growthRate, projectedFullDate, err := s.calculateGrowthRate(ctx, drive.ID, usageData.UsedSize)
	if err != nil {
		slog.Warn("Failed to calculate growth rate",
			"drive_id", drive.ID,
			"error", err)
		// Continue without growth rate data
	}

	usageData.GrowthRateGBPerDay = growthRate
	usageData.ProjectedFullDate = projectedFullDate

	// Record usage data
	_, err = s.repo.RecordDriveUsage(ctx, repository.RecordDriveUsageParams{
		DriveID:            drive.ID,
		TotalSize:          usageData.TotalSize,
		UsedSize:           usageData.UsedSize,
		AvailableSize:      usageData.AvailableSize,
		UsagePercentage:    usageData.UsagePercentage,
		GrowthRateGbPerDay: sql.NullFloat64{Float64: growthRate, Valid: growthRate > 0},
		ProjectedFullDate:  sql.NullTime{Time: timeFromDatePtr(projectedFullDate), Valid: projectedFullDate != nil},
	})
	if err != nil {
		slog.Error("Failed to record drive usage",
			"drive_id", drive.ID,
			"error", err)
		return err
	}

	// Update mounted_drives table with latest usage
	err = s.repo.UpdateMountedDriveUsage(ctx, repository.UpdateMountedDriveUsageParams{
		TotalSize:       sql.NullInt64{Int64: usageData.TotalSize, Valid: true},
		UsedSize:        sql.NullInt64{Int64: usageData.UsedSize, Valid: true},
		AvailableSize:   sql.NullInt64{Int64: usageData.AvailableSize, Valid: true},
		UsagePercentage: sql.NullFloat64{Float64: usageData.UsagePercentage, Valid: true},
		IsOnline:        sql.NullBool{Bool: true, Valid: true},
		ID:              drive.ID,
	})
	if err != nil {
		slog.Error("Failed to update mounted drive usage",
			"drive_id", drive.ID,
			"error", err)
		return err
	}

	// Check for alerts
	if err := s.checkAndCreateAlerts(ctx, drive, usageData); err != nil {
		slog.Error("Failed to check alerts for drive",
			"drive_id", drive.ID,
			"error", err)
		return err
	}

	slog.Debug("Successfully monitored drive",
		"drive_id", drive.ID,
		"usage_percentage", usageData.UsagePercentage,
		"growth_rate_gb_per_day", growthRate)

	return nil
}

// getDriveUsage gets filesystem usage statistics for a mount path
func (s *DriveMonitorService) getDriveUsage(mountPath string) (*DriveUsageData, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(mountPath, &stat); err != nil {
		return nil, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Calculate sizes in bytes
	totalSize := int64(stat.Blocks) * int64(stat.Bsize)
	availableSize := int64(stat.Bavail) * int64(stat.Bsize)
	usedSize := totalSize - availableSize

	// Calculate usage percentage
	var usagePercentage float64
	if totalSize > 0 {
		usagePercentage = (float64(usedSize) / float64(totalSize)) * 100
	}

	return &DriveUsageData{
		TotalSize:       totalSize,
		UsedSize:        usedSize,
		AvailableSize:   availableSize,
		UsagePercentage: usagePercentage,
		RecordedAt:      time.Now(),
	}, nil
}

// calculateGrowthRate calculates the daily growth rate based on historical usage data
func (s *DriveMonitorService) calculateGrowthRate(ctx context.Context, driveID string, currentUsedSize int64) (float64, *string, error) {
	// Get usage history for the past 7 days
	history, err := s.repo.GetDriveUsageHistory(ctx, repository.GetDriveUsageHistoryParams{
		DriveID: driveID,
		Limit:   100,
	})
	if err != nil {
		return 0, nil, err
	}

	if len(history) < 2 {
		// Not enough data to calculate growth rate
		return 0, nil, nil
	}

	// Calculate average daily growth rate using linear regression or simple average
	var totalGrowthGB float64
	var validDataPoints int

	for i := 1; i < len(history); i++ {
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
		sizeDiffGB := float64(sizeDiffBytes) / (1024 * 1024 * 1024) // Convert to GB
		dailyGrowth := sizeDiffGB / (timeDiff.Hours() / 24)         // GB per day

		if dailyGrowth >= 0 { // Only consider positive growth
			totalGrowthGB += dailyGrowth
			validDataPoints++
		}
	}

	if validDataPoints == 0 {
		return 0, nil, nil
	}

	avgGrowthRateGBPerDay := totalGrowthGB / float64(validDataPoints)

	// Calculate projected full date if growth rate is meaningful
	var projectedFullDate *string
	if avgGrowthRateGBPerDay > 0.1 { // Only if growing more than 0.1GB per day
		// Get the latest drive info to know total size
		latestHistory := history[0]
		availableGB := float64(latestHistory.AvailableSize) / (1024 * 1024 * 1024)
		daysUntilFull := availableGB / avgGrowthRateGBPerDay

		if daysUntilFull > 0 && daysUntilFull < 365*2 { // Only if less than 2 years
			fullDate := time.Now().AddDate(0, 0, int(daysUntilFull))
			projectedDateStr := fullDate.Format("2006-01-02")
			projectedFullDate = &projectedDateStr
		}
	}

	return avgGrowthRateGBPerDay, projectedFullDate, nil
}

// checkAndCreateAlerts checks if any alerts should be created for the drive
func (s *DriveMonitorService) checkAndCreateAlerts(ctx context.Context, drive repository.MountedDrife, usage *DriveUsageData) error {
	// Skip monitoring if disabled for this drive
	if drive.MonitoringEnabled.Valid && !drive.MonitoringEnabled.Bool {
		slog.Debug("Drive monitoring disabled", "drive", drive.Name, "drive_id", drive.ID)
		return nil
	}

	// Use custom thresholds if available, otherwise fall back to defaults
	warningThreshold := DefaultUsageThreshold      // 80.0
	criticalThreshold := CriticalUsageThreshold    // 95.0
	growthRateThreshold := DefaultGrowthRateThreshold // 50.0

	if drive.WarningThreshold.Valid {
		warningThreshold = drive.WarningThreshold.Float64
	}
	if drive.CriticalThreshold.Valid {
		criticalThreshold = drive.CriticalThreshold.Float64
	}
	if drive.GrowthRateThreshold.Valid {
		growthRateThreshold = drive.GrowthRateThreshold.Float64
	}

	// Check usage threshold alerts
	if usage.UsagePercentage >= criticalThreshold {
		err := s.createAlert(ctx, drive, "usage_threshold", criticalThreshold, usage.UsagePercentage,
			fmt.Sprintf("CRITICAL: Drive '%s' is %s full. Immediate action required to free up space.", 
				drive.Name, formatPercentage(usage.UsagePercentage)))
		if err != nil {
			return err
		}
	} else if usage.UsagePercentage >= warningThreshold {
		err := s.createAlert(ctx, drive, "usage_threshold", warningThreshold, usage.UsagePercentage,
			fmt.Sprintf("WARNING: Drive '%s' is %s full. Consider freeing up space soon.", 
				drive.Name, formatPercentage(usage.UsagePercentage)))
		if err != nil {
			return err
		}
	}

	// Check growth rate alerts
	if usage.GrowthRateGBPerDay >= growthRateThreshold {
		err := s.createAlert(ctx, drive, "growth_rate", growthRateThreshold, usage.GrowthRateGBPerDay,
			fmt.Sprintf("Drive '%s' is growing rapidly at %.1f GB per day. Monitor closely.", 
				drive.Name, usage.GrowthRateGBPerDay))
		if err != nil {
			return err
		}
	}

	// Check projected full date alerts
	if usage.ProjectedFullDate != nil {
		projectedDate, err := time.Parse("2006-01-02", *usage.ProjectedFullDate)
		if err == nil {
			daysUntilFull := time.Until(projectedDate).Hours() / 24
			if daysUntilFull <= 30 && daysUntilFull > 0 { // Alert if full within 30 days
				err := s.createAlert(ctx, drive, "projected_full", 30, daysUntilFull,
					fmt.Sprintf("Drive '%s' is projected to be full by %s (in %.0f days).", 
						drive.Name, *usage.ProjectedFullDate, daysUntilFull))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createAlert creates a new alert if one doesn't already exist for this condition
func (s *DriveMonitorService) createAlert(ctx context.Context, drive repository.MountedDrife, alertType string, threshold, currentValue float64, message string) error {
	// Check if an active alert already exists for this drive and alert type
	existingAlerts, err := s.repo.GetActiveDriveAlerts(ctx)
	if err != nil {
		return err
	}

	for _, alert := range existingAlerts {
		if alert.DriveID == drive.ID && alert.AlertType == alertType && alert.IsActive.Valid && alert.IsActive.Bool {
			// Alert already exists, just log it
			slog.Debug("Alert already exists for drive",
				"drive_id", drive.ID,
				"alert_type", alertType,
				"current_value", currentValue)
			return nil
		}
	}

	// Create new alert
	_, err = s.repo.CreateDriveAlert(ctx, repository.CreateDriveAlertParams{
		DriveID:        drive.ID,
		AlertType:      alertType,
		ThresholdValue: threshold,
		CurrentValue:   currentValue,
		AlertMessage:   message,
	})
	if err != nil {
		return err
	}

	slog.Info("Created drive alert",
		"drive_id", drive.ID,
		"drive_name", drive.Name,
		"alert_type", alertType,
		"threshold", threshold,
		"current_value", currentValue,
		"message", message)

	return nil
}

// GetActiveDriveAlerts returns all currently active drive alerts
func (s *DriveMonitorService) GetActiveDriveAlerts(ctx context.Context) ([]DriveAlert, error) {
	alerts, err := s.repo.GetActiveDriveAlerts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]DriveAlert, len(alerts))
	for i, alert := range alerts {
		var lastTriggered string
		if alert.LastTriggered.Valid {
			lastTriggered = alert.LastTriggered.Time.Format(time.RFC3339)
		}
		
		result[i] = DriveAlert{
			ID:               alert.ID,
			DriveID:          alert.DriveID,
			DriveName:        alert.DriveName,
			MountPath:        alert.MountPath,
			AlertType:        alert.AlertType,
			ThresholdValue:   alert.ThresholdValue,
			CurrentValue:     alert.CurrentValue,
			AlertMessage:     alert.AlertMessage,
			IsActive:         alert.IsActive.Bool,
			LastTriggered:    lastTriggered,
			AcknowledgeCount: alert.AcknowledgementCount.Int64,
		}
	}

	return result, nil
}

// AcknowledgeAlert marks an alert as acknowledged (deactivates it)
func (s *DriveMonitorService) AcknowledgeAlert(ctx context.Context, alertID int64) error {
	err := s.repo.DeactivateDriveAlert(ctx, alertID)
	if err != nil {
		return err
	}

	slog.Info("Drive alert acknowledged", "alert_id", alertID)
	return nil
}

// GetDriveUsageHistory returns historical usage data for a drive
func (s *DriveMonitorService) GetDriveUsageHistory(ctx context.Context, driveID string, days int) ([]DriveUsageData, error) {
	history, err := s.repo.GetDriveUsageHistory(ctx, repository.GetDriveUsageHistoryParams{
		DriveID: driveID,
		Limit:   1000,
	})
	if err != nil {
		return nil, err
	}

	result := make([]DriveUsageData, len(history))
	for i, record := range history {
		var projectedFullDate *string
		if record.ProjectedFullDate.Valid {
			dateStr := record.ProjectedFullDate.Time.Format("2006-01-02")
			projectedFullDate = &dateStr
		}
		
		var recordedAt time.Time
		if record.RecordedAt.Valid {
			recordedAt = record.RecordedAt.Time
		}
		
		result[i] = DriveUsageData{
			TotalSize:          record.TotalSize,
			UsedSize:           record.UsedSize,
			AvailableSize:      record.AvailableSize,
			UsagePercentage:    record.UsagePercentage,
			GrowthRateGBPerDay: record.GrowthRateGbPerDay.Float64,
			ProjectedFullDate:  projectedFullDate,
			RecordedAt:         recordedAt,
		}
	}

	return result, nil
}

// Utility functions
func formatPercentage(percentage float64) string {
	return fmt.Sprintf("%.1f%%", percentage)
}

func timeFromDatePtr(dateStr *string) time.Time {
	if dateStr == nil {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}