package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/pkg/structures"
)

// Manager coordinates all job execution
type Manager struct {
	gctx     global.Context
	jobs     map[structures.Job]Job
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewManager creates a new job manager
func NewManager(gctx global.Context) *Manager {
	return &Manager{
		gctx:     gctx,
		jobs:     make(map[structures.Job]Job),
		stopChan: make(chan struct{}),
	}
}

// Register registers a job with the manager
func (m *Manager) Register(job Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := job.Name()
	if _, exists := m.jobs[name]; exists {
		return fmt.Errorf("job %s already registered", name)
	}

	m.jobs[name] = job
	slog.Info("Registered job", "name", name)
	return nil
}

// RegisterAll registers multiple jobs by name
func (m *Manager) RegisterAll(names ...structures.Job) error {
	for _, name := range names {
		job, err := NewJob(name, m.gctx)
		if err != nil {
			return fmt.Errorf("failed to create job %s: %w", name, err)
		}
		if err := m.Register(job); err != nil {
			return fmt.Errorf("failed to register job %s: %w", name, err)
		}
	}
	return nil
}

// Start starts all registered jobs
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("job manager already running")
	}

	slog.Info("Starting job manager", "job_count", len(m.jobs))
	m.running = true

	// Start each job
	for name, job := range m.jobs {
		if job.Config().Enabled {
			m.wg.Add(1)
			go m.runJob(ctx, job)
			slog.Info("Started job", "name", name)
		} else {
			slog.Info("Skipping disabled job", "name", name)
		}
	}

	return nil
}

// Stop stops all jobs gracefully
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	slog.Info("Stopping job manager")
	m.running = false
	close(m.stopChan)

	// Stop all jobs
	for name, job := range m.jobs {
		if err := job.Stop(ctx); err != nil {
			slog.Error("Failed to stop job", "name", name, "error", err)
		}
	}

	// Wait for all jobs to complete with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("All jobs stopped successfully")
	case <-ctx.Done():
		slog.Warn("Timeout waiting for jobs to stop")
		return ctx.Err()
	}

	return nil
}

// runJob manages the lifecycle of a single job
func (m *Manager) runJob(ctx context.Context, job Job) {
	defer m.wg.Done()

	name := job.Name()
	config := job.Config()

	slog.Debug("Starting job runner", "name", name, "interval", config.Interval)

	// Start the job
	if err := job.Start(ctx); err != nil {
		slog.Error("Failed to start job", "name", name, "error", err)
		job.OnError(ctx, err)
		return
	}

	// Run on startup if configured
	if config.RunOnStartup {
		m.executeJob(ctx, job)
	}

	// Set up ticker for periodic execution
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.executeJob(ctx, job)
		case <-m.stopChan:
			slog.Debug("Job runner stopping", "name", name)
			return
		case <-ctx.Done():
			slog.Debug("Job runner context cancelled", "name", name)
			return
		}
	}
}

// executeJob executes a job with timeout and retry logic
func (m *Manager) executeJob(ctx context.Context, job Job) {
	name := job.Name()
	config := job.Config()
	start := time.Now()

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			slog.Warn("Retrying job execution", "name", name, "attempt", attempt, "max_retries", config.MaxRetries)
			
			// Wait before retry
			select {
			case <-time.After(config.RetryDelay):
			case <-execCtx.Done():
				job.OnError(execCtx, execCtx.Err())
				return
			}
		}

		// Execute the job
		err := job.Trigger(execCtx)
		if err == nil {
			duration := time.Since(start)
			slog.Debug("Job executed successfully", "name", name, "duration", duration)
			job.OnSuccess(execCtx, duration)
			return
		}

		lastErr = err
		slog.Warn("Job execution failed", "name", name, "attempt", attempt+1, "error", err)
	}

	// All retries exhausted
	slog.Error("Job failed after all retries", "name", name, "error", lastErr)
	job.OnError(execCtx, lastErr)
}

// GetJob returns a specific job by name
func (m *Manager) GetJob(name structures.Job) (Job, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, exists := m.jobs[name]
	return job, exists
}

// ListJobs returns all registered jobs
func (m *Manager) ListJobs() []Job {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	jobs := make([]Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetMetrics returns metrics for all jobs
func (m *Manager) GetMetrics() map[structures.Job]JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	metrics := make(map[structures.Job]JobMetrics)
	for name, job := range m.jobs {
		metrics[name] = job.Metrics()
	}
	return metrics
}

// TriggerJob manually triggers a specific job
func (m *Manager) TriggerJob(ctx context.Context, name structures.Job) error {
	job, exists := m.GetJob(name)
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	go m.executeJob(ctx, job)
	return nil
}

// UpdateJobConfig updates configuration for a specific job
func (m *Manager) UpdateJobConfig(name structures.Job, config JobConfig) error {
	job, exists := m.GetJob(name)
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	return job.SetConfig(config)
}

// HealthCheck returns the health status of all jobs
func (m *Manager) HealthCheck() map[structures.Job]error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	health := make(map[structures.Job]error)
	for name, job := range m.jobs {
		health[name] = job.Health()
	}
	return health
}

// IsRunning returns whether the manager is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}