package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// InvitationCleanup job removes expired invitations from the database
type InvitationCleanup struct {
	*BaseJob
	gctx global.Context
}

// NewInvitationCleanup creates a new invitation cleanup job
func NewInvitationCleanup(gctx global.Context, config JobConfig) (Job, error) {
	baseJob := NewBaseJob(gctx, structures.JobInvitationCleanup, config)
	
	return &InvitationCleanup{
		BaseJob: baseJob,
		gctx:    gctx,
	}, nil
}

// Name returns the job name
func (j *InvitationCleanup) Name() structures.Job {
	return structures.JobInvitationCleanup
}

// Trigger executes the invitation cleanup task
func (j *InvitationCleanup) Trigger(ctx context.Context) error {
	start := time.Now()
	
	slog.Info("Starting invitation cleanup job")
	
	// Call the ExpireOldInvitations function
	err := j.gctx.Crate().Sqlite.Query().ExpireOldInvitations(ctx)
	if err != nil {
		slog.Error("Failed to expire old invitations", "error", err)
		return err
	}
	
	duration := time.Since(start)
	
	slog.Info("Invitation cleanup completed successfully", "duration", duration)
	
	return nil
}

// Start initializes the job
func (j *InvitationCleanup) Start(ctx context.Context) error {
	slog.Info("Invitation cleanup job started")
	return nil
}

// Stop cleans up the job
func (j *InvitationCleanup) Stop(ctx context.Context) error {
	slog.Info("Invitation cleanup job stopped")
	return nil
}

// Health returns the job health status
func (j *InvitationCleanup) Health() error {
	return nil // Simple job, always healthy if running
}