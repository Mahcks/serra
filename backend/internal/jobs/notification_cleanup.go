package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// NotificationCleanup job removes expired notifications from the database
type NotificationCleanup struct {
	*BaseJob
	gctx global.Context
}

// NewNotificationCleanup creates a new notification cleanup job
func NewNotificationCleanup(gctx global.Context, config JobConfig) (Job, error) {
	baseJob := NewBaseJob(gctx, structures.JobNotificationCleanup, config)
	
	return &NotificationCleanup{
		BaseJob: baseJob,
		gctx:    gctx,
	}, nil
}

// Name returns the job name
func (j *NotificationCleanup) Name() structures.Job {
	return structures.JobNotificationCleanup
}

// Trigger executes the notification cleanup task
func (j *NotificationCleanup) Trigger(ctx context.Context) error {
	start := time.Now()
	
	slog.Info("Starting notification cleanup job")
	
	// Clean up expired notifications
	err := j.gctx.Crate().NotificationService.CleanupExpiredNotifications(ctx)
	if err != nil {
		slog.Error("Failed to cleanup expired notifications", "error", err)
		return err
	}
	
	duration := time.Since(start)
	
	slog.Info("Notification cleanup completed successfully", "duration", duration)
	
	return nil
}

// Start initializes the job
func (j *NotificationCleanup) Start(ctx context.Context) error {
	slog.Info("Notification cleanup job started")
	return nil
}

// Stop cleans up the job
func (j *NotificationCleanup) Stop(ctx context.Context) error {
	slog.Info("Notification cleanup job stopped")
	return nil
}

// Health returns the job health status
func (j *NotificationCleanup) Health() error {
	return nil // Simple job, always healthy if running
}