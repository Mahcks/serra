package jobs

import (
	"context"
	"time"

	"github.com/mahcks/serra/pkg/structures"
)

// JobStatus represents the current status of a job
type JobStatus string

const (
	JobStatusStopped  JobStatus = "stopped"
	JobStatusRunning  JobStatus = "running"
	JobStatusError    JobStatus = "error"
	JobStatusStopping JobStatus = "stopping"
)

// JobMetrics provides metrics about job execution
type JobMetrics struct {
	Name            structures.Job `json:"name"`
	Status          JobStatus      `json:"status"`
	LastRun         time.Time      `json:"last_run"`
	NextRun         *time.Time     `json:"next_run,omitempty"`
	RunCount        int64          `json:"run_count"`
	ErrorCount      int64          `json:"error_count"`
	AverageRunTime  time.Duration  `json:"average_run_time"`
	LastError       string         `json:"last_error,omitempty"`
	LastErrorTime   *time.Time     `json:"last_error_time,omitempty"`
}

// JobConfig provides configuration for a job
type JobConfig struct {
	Enabled      bool          `json:"enabled"`
	Interval     time.Duration `json:"interval"`
	MaxRetries   int           `json:"max_retries"`
	RetryDelay   time.Duration `json:"retry_delay"`
	Timeout      time.Duration `json:"timeout"`
	RunOnStartup bool          `json:"run_on_startup"`
}

// Job is the interface all jobs must implement.
type Job interface {
	// Lifecycle methods
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	
	// Identity and configuration
	Name() structures.Job
	Config() JobConfig
	SetConfig(config JobConfig) error
	
	// Execution
	Trigger(ctx context.Context) error
	
	// Monitoring
	Status() JobStatus
	Metrics() JobMetrics
	Health() error
	
	// Events (optional, can return nil if not implemented)
	OnError(ctx context.Context, err error)
	OnSuccess(ctx context.Context, duration time.Duration)
}
