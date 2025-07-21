package analytics

import (
	"log/slog"
	
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/storage_pools"
	"github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetAnalyticsOverview returns a comprehensive overview of system analytics
func (rg *RouteGroup) GetAnalyticsOverview(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	// Get active drive alerts
	alerts, err := rg.driveMonitor.GetActiveDriveAlerts(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get drive alerts")
	}

	// Get all mounted drives first
	allMountedDrives, err := rg.gctx.Crate().Sqlite.Query().ListMountedDrives(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get mounted drives")
	}
	
	// Debug logging
	slog.Info("Analytics: Found mounted drives", "count", len(allMountedDrives))
	for i, drive := range allMountedDrives {
		slog.Info("Analytics: Mounted drive", "index", i, "id", drive.ID, "name", drive.Name, "mount_path", drive.MountPath, "is_online", drive.IsOnline.Bool)
	}

	// Get latest drive usage for monitored drives
	latestUsage, err := rg.gctx.Crate().Sqlite.Query().GetLatestDriveUsage(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get latest drive usage")
	}
	
	slog.Info("Analytics: Found usage records", "count", len(latestUsage))
	for i, usage := range latestUsage {
		slog.Info("Analytics: Usage record", "index", i, "drive_id", usage.DriveID, "usage_percentage", usage.UsagePercentage)
	}

	// Get request analytics (top 10 most requested content)
	requestAnalytics, err := rg.gctx.Crate().Sqlite.Query().GetRequestAnalytics(ctx.Context(), 10)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get request analytics")
	}

	// Get trending content
	trendingContent, err := rg.gctx.Crate().Sqlite.Query().GetTrendingContent(ctx.Context(), 10)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get trending content")
	}

	// Create a map of usage data by drive ID for quick lookup
	usageByDriveID := make(map[string]interface{})
	for _, usage := range latestUsage {
		usageByDriveID[usage.DriveID] = map[string]interface{}{
			"drive_id":            usage.DriveID,
			"name":                usage.DriveName,
			"mount_path":          usage.DriveMountPath,
			"total_size":          usage.TotalSize,
			"used_size":           usage.UsedSize,
			"available_size":      usage.AvailableSize,
			"usage_percentage":    usage.UsagePercentage,
			"growth_rate_gb_per_day": usage.GrowthRateGbPerDay.Float64,
			"projected_full_date": nil,
			"recorded_at":         usage.RecordedAt.Time,
			"has_data":           true,
		}
		if usage.ProjectedFullDate.Valid {
			usageByDriveID[usage.DriveID].(map[string]interface{})["projected_full_date"] = usage.ProjectedFullDate.Time
		}
	}

	// Create combined drive list including all mounted drives
	var combinedDriveUsage []interface{}
	
	// Calculate summary statistics
	var totalDrives, criticalDrives, warningDrives int
	var totalStorage, usedStorage int64

	for _, drive := range allMountedDrives {
		totalDrives++
		
		if usageData, exists := usageByDriveID[drive.ID]; exists {
			// Drive has usage data - already includes metadata from the joined query
			usageMap := usageData.(map[string]interface{})
			
			// Add online status and threshold values from mounted_drives table
			usageMap["is_online"] = drive.IsOnline.Bool
			usageMap["monitoring_enabled"] = drive.MonitoringEnabled.Bool
			usageMap["warning_threshold"] = drive.WarningThreshold.Float64
			usageMap["critical_threshold"] = drive.CriticalThreshold.Float64
			usageMap["growth_rate_threshold"] = drive.GrowthRateThreshold.Float64
			
			combinedDriveUsage = append(combinedDriveUsage, usageMap)
			
			totalStorage += usageMap["total_size"].(int64)
			usedStorage += usageMap["used_size"].(int64)
			
			usagePercent := usageMap["usage_percentage"].(float64)
			if usagePercent >= 95 {
				criticalDrives++
			} else if usagePercent >= 80 {
				warningDrives++
			}
		} else {
			// Drive exists but no usage data (offline or not monitored)
			combinedDriveUsage = append(combinedDriveUsage, map[string]interface{}{
				"drive_id":             drive.ID,
				"name":                 drive.Name,
				"mount_path":           drive.MountPath,
				"total_size":           0,
				"used_size":            0,
				"available_size":       0,
				"usage_percentage":     0,
				"growth_rate_gb_per_day": 0,
				"projected_full_date":  nil,
				"recorded_at":          nil,
				"has_data":            false,
				"is_online":           drive.IsOnline.Bool,
				"monitoring_enabled":   drive.MonitoringEnabled.Bool,
				"warning_threshold":    drive.WarningThreshold.Float64,
				"critical_threshold":   drive.CriticalThreshold.Float64,
				"growth_rate_threshold": drive.GrowthRateThreshold.Float64,
			})
		}
	}

	var overallUsagePercentage float64
	if totalStorage > 0 {
		overallUsagePercentage = (float64(usedStorage) / float64(totalStorage)) * 100
	}

	// Get storage pools information
	poolService := storage_pools.NewStoragePoolService()
	storagePools, err := poolService.DetectAllPools()
	if err != nil {
		slog.Warn("Failed to detect storage pools", "error", err)
		storagePools = []structures.StoragePool{} // Empty slice if detection fails
	}

	// Debug the final result
	slog.Info("Analytics: Final result", "total_drives", totalDrives, "combined_drive_usage_count", len(combinedDriveUsage), "storage_pools_count", len(storagePools))

	return ctx.JSON(map[string]interface{}{
		"summary": map[string]interface{}{
			"total_drives":            totalDrives,
			"drives_with_alerts":      len(alerts),
			"critical_drives":         criticalDrives,
			"warning_drives":          warningDrives,
			"overall_usage_percent":   overallUsagePercentage,
			"total_storage_gb":        float64(totalStorage) / (1024 * 1024 * 1024),
			"used_storage_gb":         float64(usedStorage) / (1024 * 1024 * 1024),
			"available_storage_gb":    float64(totalStorage-usedStorage) / (1024 * 1024 * 1024),
		},
		"active_alerts":     alerts,
		"drive_usage":       combinedDriveUsage,
		"storage_pools":     storagePools,
		"request_analytics": requestAnalytics,
		"trending_content":  trendingContent,
	})
}