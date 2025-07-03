package jobs

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// BaseJob provides common functionality for all jobs
type BaseJob struct {
	gctx          global.Context
	name          structures.Job
	config        JobConfig
	status        int32 // atomic: 0=stopped, 1=running, 2=error, 3=stopping
	
	// Metrics (atomic for thread safety)
	runCount       int64
	errorCount     int64
	totalRunTime   int64 // nanoseconds
	lastRun        int64 // unix nano
	lastError      string
	lastErrorTime  int64 // unix nano
	
	mu            sync.RWMutex
	stopChan      chan struct{}
	running       bool
}

// NewBaseJob creates a new base job
func NewBaseJob(gctx global.Context, name structures.Job, config JobConfig) *BaseJob {
	return &BaseJob{
		gctx:     gctx,
		name:     name,
		config:   config,
		status:   statusToInt32(JobStatusStopped),
		stopChan: make(chan struct{}),
	}
}

// Name returns the job name
func (b *BaseJob) Name() structures.Job {
	return b.name
}

// Config returns the job configuration
func (b *BaseJob) Config() JobConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// SetConfig updates the job configuration
func (b *BaseJob) SetConfig(config JobConfig) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.config = config
	return nil
}

// statusToInt32 converts JobStatus to int32 for atomic operations
func statusToInt32(status JobStatus) int32 {
	switch status {
	case JobStatusStopped:
		return 0
	case JobStatusRunning:
		return 1
	case JobStatusError:
		return 2
	case JobStatusStopping:
		return 3
	default:
		return 0
	}
}

// int32ToStatus converts int32 back to JobStatus
func int32ToStatus(val int32) JobStatus {
	switch val {
	case 0:
		return JobStatusStopped
	case 1:
		return JobStatusRunning
	case 2:
		return JobStatusError
	case 3:
		return JobStatusStopping
	default:
		return JobStatusStopped
	}
}

// Status returns the current job status
func (b *BaseJob) Status() JobStatus {
	return int32ToStatus(atomic.LoadInt32(&b.status))
}

// setStatus atomically sets the job status
func (b *BaseJob) setStatus(status JobStatus) {
	atomic.StoreInt32(&b.status, statusToInt32(status))
}

// Start starts the base job (should be called by concrete implementations)
func (b *BaseJob) Start(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.running {
		return nil
	}
	
	b.running = true
	b.setStatus(JobStatusRunning)
	return nil
}

// Stop stops the base job
func (b *BaseJob) Stop(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if !b.running {
		return nil
	}
	
	b.setStatus(JobStatusStopping)
	close(b.stopChan)
	b.running = false
	b.setStatus(JobStatusStopped)
	return nil
}

// IsRunning returns whether the job is currently running
func (b *BaseJob) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// ShouldStop returns whether the job should stop
func (b *BaseJob) ShouldStop() bool {
	select {
	case <-b.stopChan:
		return true
	default:
		return false
	}
}

// Metrics returns job execution metrics
func (b *BaseJob) Metrics() JobMetrics {
	runCount := atomic.LoadInt64(&b.runCount)
	errorCount := atomic.LoadInt64(&b.errorCount)
	totalRunTime := time.Duration(atomic.LoadInt64(&b.totalRunTime))
	lastRunNano := atomic.LoadInt64(&b.lastRun)
	lastErrorTimeNano := atomic.LoadInt64(&b.lastErrorTime)
	
	metrics := JobMetrics{
		Name:       b.name,
		Status:     b.Status(),
		RunCount:   runCount,
		ErrorCount: errorCount,
	}
	
	if lastRunNano > 0 {
		metrics.LastRun = time.Unix(0, lastRunNano)
	}
	
	if runCount > 0 {
		metrics.AverageRunTime = time.Duration(totalRunTime.Nanoseconds() / runCount)
	}
	
	if lastErrorTimeNano > 0 {
		lastErrorTime := time.Unix(0, lastErrorTimeNano)
		metrics.LastErrorTime = &lastErrorTime
		
		b.mu.RLock()
		metrics.LastError = b.lastError
		b.mu.RUnlock()
	}
	
	// Calculate next run based on interval and last run
	config := b.Config()
	if lastRunNano > 0 && config.Enabled {
		nextRun := time.Unix(0, lastRunNano).Add(config.Interval)
		metrics.NextRun = &nextRun
	}
	
	return metrics
}

// Health performs a basic health check
func (b *BaseJob) Health() error {
	// Base implementation just checks if job is responsive
	// Concrete implementations can override this
	if b.Status() == JobStatusError {
		b.mu.RLock()
		lastError := b.lastError
		b.mu.RUnlock()
		return fmt.Errorf("job in error state: %s", lastError)
	}
	return nil
}

// OnSuccess records successful execution
func (b *BaseJob) OnSuccess(ctx context.Context, duration time.Duration) {
	atomic.AddInt64(&b.runCount, 1)
	atomic.AddInt64(&b.totalRunTime, duration.Nanoseconds())
	atomic.StoreInt64(&b.lastRun, time.Now().UnixNano())
	
	// Clear error state if we were in error
	if b.Status() == JobStatusError {
		b.setStatus(JobStatusRunning)
		b.mu.Lock()
		b.lastError = ""
		b.mu.Unlock()
		atomic.StoreInt64(&b.lastErrorTime, 0)
	}
}

// OnError records failed execution
func (b *BaseJob) OnError(ctx context.Context, err error) {
	atomic.AddInt64(&b.errorCount, 1)
	atomic.StoreInt64(&b.lastErrorTime, time.Now().UnixNano())
	
	b.mu.Lock()
	b.lastError = err.Error()
	b.mu.Unlock()
	
	b.setStatus(JobStatusError)
}

// Context returns the global context
func (b *BaseJob) Context() global.Context {
	return b.gctx
}