package jobs

import (
	"fmt"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// Default job configurations
var defaultConfigs = map[structures.Job]JobConfig{
	structures.JobDownloadPoller: {
		Enabled:      true,
		Interval:     30 * time.Second,
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
		Timeout:     1 * time.Minute,
		RunOnStartup: false,
	},
}

// NewJob creates a job by name with default configuration
func NewJob(name structures.Job, gctx global.Context) (Job, error) {
	config, exists := defaultConfigs[name]
	if !exists {
		return nil, fmt.Errorf("unknown job: %s", name)
	}

	switch name {
	case structures.JobDownloadPoller:
		return NewDownloadPoller(gctx, config)
	case structures.JobDriveMonitor:
		return NewDriveMonitor(gctx, config)
	default:
		return nil, fmt.Errorf("unknown job: %s", name)
	}
}

// NewJobWithConfig creates a job with custom configuration
func NewJobWithConfig(name structures.Job, gctx global.Context, config JobConfig) (Job, error) {
	switch name {
	case structures.JobDownloadPoller:
		return NewDownloadPoller(gctx, config)
	case structures.JobDriveMonitor:
		return NewDriveMonitor(gctx, config)
	default:
		return nil, fmt.Errorf("unknown job: %s", name)
	}
}

// AllJobNames returns all available job names
func AllJobNames() []structures.Job {
	return []structures.Job{structures.JobDownloadPoller, structures.JobDriveMonitor}
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
