package jobs

import (
	"fmt"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/pkg/structures"
)

// Default job configurations
var defaultConfigs = map[structures.Job]JobConfig{
	structures.JobDownloadPoller: {
		Enabled:      true,
		Interval:     10 * time.Second,
		MaxRetries:   3,
		RetryDelay:   5 * time.Second,
		Timeout:      2 * time.Minute,
		RunOnStartup: true,
	},
	structures.JobDriveMonitor: {
		Enabled:      true,
		Interval:     5 * time.Minute,
		MaxRetries:   2,
		RetryDelay:   30 * time.Second,
		Timeout:      1 * time.Minute,
		RunOnStartup: false,
	},
	structures.JobRequestProcessor: {
		Enabled:      true,
		Interval:     20 * time.Second,
		MaxRetries:   3,
		RetryDelay:   30 * time.Second,
		Timeout:      1 * time.Minute,
		RunOnStartup: false,
	},
	structures.JobLibrarySyncFull: {
		Enabled:      false,          // Disabled by default, only enabled in dev
		Interval:     24 * time.Hour, // Full sync every 24 hours
		MaxRetries:   2,
		RetryDelay:   10 * time.Minute,
		Timeout:      15 * time.Minute,
		RunOnStartup: false, // Don't run on startup by default
	},
	structures.JobLibrarySyncIncremental: {
		Enabled:      false,            // Disabled by default, only enabled in dev
		Interval:     15 * time.Minute, // Incremental sync every 15 minutes
		MaxRetries:   3,
		RetryDelay:   2 * time.Minute,
		Timeout:      5 * time.Minute,
		RunOnStartup: false, // Don't run on startup
	},
	structures.JobInvitationCleanup: {
		Enabled:      true,
		Interval:     1 * time.Hour, // Clean up expired invitations every hour
		MaxRetries:   2,
		RetryDelay:   10 * time.Minute,
		Timeout:      30 * time.Second,
		RunOnStartup: false, // Don't run on startup
	},
	structures.JobNotificationCleanup: {
		Enabled:      true,
		Interval:     1 * time.Hour, // Clean up expired notifications every hour
		MaxRetries:   2,
		RetryDelay:   10 * time.Minute,
		Timeout:      30 * time.Second,
		RunOnStartup: false, // Don't run on startup
	},
}

// NewJob creates a job by name with default configuration
func NewJob(name structures.Job, gctx global.Context, integrations *integrations.Integration) (Job, error) {
	config, exists := defaultConfigs[name]
	if !exists {
		return nil, fmt.Errorf("unknown job: %s", name)
	}

	switch name {
	case structures.JobDownloadPoller:
		return NewDownloadPoller(gctx, config)
	case structures.JobDriveMonitor:
		return NewDriveMonitor(gctx, config)
	case structures.JobRequestProcessor:
		return NewRequestProcessor(gctx, integrations, config)
	case structures.JobLibrarySyncFull:
		// Only enable library sync in dev mode
		if gctx.Bootstrap().Version == "dev" {
			config.Enabled = true
			config.RunOnStartup = false
		}
		return NewLibrarySyncFull(gctx, config)
	case structures.JobLibrarySyncIncremental:
		// Only enable library sync in dev mode
		if gctx.Bootstrap().Version == "dev" {
			config.Enabled = false
		}
		return NewLibrarySyncIncremental(gctx, config)
	case structures.JobInvitationCleanup:
		return NewInvitationCleanup(gctx, config)
	case structures.JobNotificationCleanup:
		return NewNotificationCleanup(gctx, config)
	default:
		return nil, fmt.Errorf("unknown job: %s", name)
	}
}

// NewJobWithConfig creates a job with custom configuration
func NewJobWithConfig(name structures.Job, gctx global.Context, integrations *integrations.Integration, config JobConfig) (Job, error) {
	switch name {
	case structures.JobDownloadPoller:
		return NewDownloadPoller(gctx, config)
	case structures.JobDriveMonitor:
		return NewDriveMonitor(gctx, config)
	case structures.JobRequestProcessor:
		return NewRequestProcessor(gctx, integrations, config)
	case structures.JobLibrarySyncFull:
		// Only enable library sync in dev mode
		if gctx.Bootstrap().Version == "dev" {
			config.Enabled = true
			config.RunOnStartup = true
		}
		return NewLibrarySyncFull(gctx, config)
	case structures.JobLibrarySyncIncremental:
		// Only enable library sync in dev mode
		if gctx.Bootstrap().Version == "dev" {
			config.Enabled = true
		}
		return NewLibrarySyncIncremental(gctx, config)
	case structures.JobInvitationCleanup:
		return NewInvitationCleanup(gctx, config)
	case structures.JobNotificationCleanup:
		return NewNotificationCleanup(gctx, config)
	default:
		return nil, fmt.Errorf("unknown job: %s", name)
	}
}

// AllJobNames returns all available job names
func AllJobNames() []structures.Job {
	return []structures.Job{structures.JobDownloadPoller, structures.JobDriveMonitor, structures.JobRequestProcessor, structures.JobLibrarySyncFull, structures.JobLibrarySyncIncremental, structures.JobInvitationCleanup, structures.JobNotificationCleanup}
}

// GetDefaultConfig returns the default configuration for a job
func GetDefaultConfig(name structures.Job) (JobConfig, bool) {
	config, exists := defaultConfigs[name]
	return config, exists
}

// SetDefaultConfig updates the default configuration for a job type
func SetDefaultConfig(name structures.Job, config JobConfig) {
	defaultConfigs[name] = config
}
