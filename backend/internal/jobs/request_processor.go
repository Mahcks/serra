package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/services/request_processor"
	"github.com/mahcks/serra/pkg/structures"
)

type RequestProcessorJob struct {
	*BaseJob
	processor request_processor.Service
}

func NewRequestProcessor(gctx global.Context, integrations *integrations.Integration, config JobConfig) (*RequestProcessorJob, error) {
	// Initialize Radarr and Sonarr services
	radarrSvc := radarr.New(gctx.Crate().Sqlite.Query())
	sonarrSvc := sonarr.New(gctx.Crate().Sqlite.Query())

	// Initialize request processor with integrations
	processor := request_processor.New(gctx.Crate().Sqlite.Query(), radarrSvc, sonarrSvc, integrations)

	base := NewBaseJob(gctx, structures.JobRequestProcessor, config)
	job := &RequestProcessorJob{
		BaseJob:   base,
		processor: processor,
	}

	return job, nil
}

// Trigger executes the request processing task
func (j *RequestProcessorJob) Trigger(ctx context.Context) error {
	return j.Execute(ctx)
}

// Start begins the request processor loop
func (j *RequestProcessorJob) Start(ctx context.Context) error {
	slog.Info("Starting request processor", "interval", j.Config().Interval)
	return j.BaseJob.Start(ctx)
}

func (j *RequestProcessorJob) Execute(ctx context.Context) error {
	slog.Debug("Starting request processor job")

	// Get all approved requests that haven't been fulfilled
	requests, err := j.Context().Crate().Sqlite.Query().GetRequestsByStatus(ctx, "approved")
	if err != nil {
		return fmt.Errorf("failed to get approved requests: %w", err)
	}

	slog.Debug("Found approved requests to check", "count", len(requests))

	processedCount := 0
	for _, req := range requests {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if this request's media has been downloaded
		err := j.processor.CheckRequestStatus(ctx, req.ID)
		if err != nil {
			slog.Error("Failed to check request status",
				"request_id", req.ID,
				"title", req.Title,
				"error", err)
			continue
		}
		processedCount++
	}

	slog.Debug("Request processor job completed",
		"total_requests", len(requests),
		"processed", processedCount)

	return nil
}
