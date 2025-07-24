package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/services/request_processor"
)

type RequestRetryJob struct {
	queries          *repository.Queries
	requestProcessor request_processor.Service
	stopChan         chan struct{}
	retryInterval    time.Duration
}

func NewRequestRetryJob(queries *repository.Queries, integrations *integrations.Integration) *RequestRetryJob {
	// Create the request processor service
	processor := request_processor.New(queries, nil, nil, integrations)
	
	return &RequestRetryJob{
		queries:          queries,
		requestProcessor: processor,
		stopChan:         make(chan struct{}),
		retryInterval:    10 * time.Minute, // Retry every 10 minutes
	}
}

func (j *RequestRetryJob) Start(ctx context.Context) {
	slog.Info("Starting request retry job", "interval", j.retryInterval)
	
	ticker := time.NewTicker(j.retryInterval)
	defer ticker.Stop()

	// Run once immediately
	j.retryFailedRequests(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Request retry job stopped due to context cancellation")
			return
		case <-j.stopChan:
			slog.Info("Request retry job stopped")
			return
		case <-ticker.C:
			j.retryFailedRequests(ctx)
		}
	}
}

func (j *RequestRetryJob) Stop() {
	close(j.stopChan)
}

func (j *RequestRetryJob) retryFailedRequests(ctx context.Context) {
	slog.Debug("Checking for failed requests to retry")
	
	// Get all failed requests
	failedRequests, err := j.queries.GetRequestsByStatus(ctx, "failed")
	if err != nil {
		slog.Error("Failed to get failed requests", "error", err)
		return
	}

	if len(failedRequests) == 0 {
		slog.Debug("No failed requests found")
		return
	}

	slog.Info("Found failed requests to retry", "count", len(failedRequests))

	for _, request := range failedRequests {
		// Check if request was failed recently (within last hour) to avoid spam retries
		if time.Since(request.UpdatedAt) < time.Hour {
			slog.Debug("Skipping recently failed request", 
				"request_id", request.ID, 
				"title", request.Title,
				"failed_at", request.UpdatedAt)
			continue
		}

		slog.Info("Retrying failed request", 
			"request_id", request.ID, 
			"title", request.Title,
			"media_type", request.MediaType,
			"failed_at", request.UpdatedAt)
		
		// Reset status to approved to trigger processing
		_, err := j.queries.UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
			Status: "approved",
			ID:     request.ID,
		})
		if err != nil {
			slog.Error("Failed to reset failed request status", 
				"request_id", request.ID, 
				"error", err)
			continue
		}

		// Process the request
		if err := j.requestProcessor.ProcessApprovedRequest(ctx, request.ID); err != nil {
			slog.Error("Failed to retry request", 
				"request_id", request.ID, 
				"title", request.Title,
				"error", err)
			// The processor will mark it as failed again
		} else {
			slog.Info("Successfully retried request", 
				"request_id", request.ID, 
				"title", request.Title)
		}

		// Small delay between retries to avoid overwhelming external services
		time.Sleep(2 * time.Second)
	}
}