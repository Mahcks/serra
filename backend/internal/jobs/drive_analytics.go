package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/services/drive_monitor"
	"github.com/mahcks/serra/pkg/structures"
)

type DriveAnalyticsJob struct {
	*BaseJob
	driveMonitor *drive_monitor.DriveMonitorService
}

func NewDriveAnalyticsJob(gctx global.Context, driveMonitor *drive_monitor.DriveMonitorService, config JobConfig) *DriveAnalyticsJob {
	base := NewBaseJob(gctx, structures.JobDriveAnalytics, config)
	return &DriveAnalyticsJob{
		BaseJob:      base,
		driveMonitor: driveMonitor,
	}
}

func (j *DriveAnalyticsJob) Trigger(ctx context.Context) error {
	slog.Info("Starting drive analytics and monitoring job")
	startTime := time.Now()

	// Monitor all drives and create alerts if needed
	if err := j.driveMonitor.MonitorAllDrives(ctx); err != nil {
		slog.Error("Failed to monitor drives", "error", err)
		return err
	}

	duration := time.Since(startTime)
	slog.Info("Drive analytics job completed successfully", 
		"duration", duration.String())

	return nil
}