package request_processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
)

type Service interface {
	ProcessApprovedRequest(ctx context.Context, requestID int64) error
	CheckRequestStatus(ctx context.Context, requestID int64) error
}

type service struct {
	repo          *repository.Queries
	radarrService radarr.Service
	sonarrService sonarr.Service
}

func New(repo *repository.Queries, radarrSvc radarr.Service, sonarrSvc sonarr.Service) Service {
	return &service{
		repo:          repo,
		radarrService: radarrSvc,
		sonarrService: sonarrSvc,
	}
}

// ProcessApprovedRequest automatically adds an approved request to Radarr/Sonarr
func (s *service) ProcessApprovedRequest(ctx context.Context, requestID int64) error {
	slog.Info("ProcessApprovedRequest called", "request_id", requestID)
	
	// Get the request details
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		slog.Error("Failed to get request by ID", "request_id", requestID, "error", err)
		return fmt.Errorf("failed to get request: %w", err)
	}

	slog.Info("Processing request", 
		"request_id", requestID,
		"title", request.Title,
		"media_type", request.MediaType, 
		"status", request.Status,
		"tmdb_id", request.TmdbID)

	if request.Status != "approved" {
		return fmt.Errorf("request %d is not approved (status: %s)", requestID, request.Status)
	}

	if !request.TmdbID.Valid {
		return fmt.Errorf("request %d has no TMDB ID", requestID)
	}

	tmdbID := request.TmdbID.Int64

	switch request.MediaType {
	case "movie":
		return s.processMovieRequest(ctx, requestID, tmdbID)
	case "tv":
		return s.processSeriesRequest(ctx, requestID, tmdbID)
	default:
		return fmt.Errorf("unsupported media type: %s", request.MediaType)
	}
}

func (s *service) processMovieRequest(ctx context.Context, requestID, tmdbID int64) error {
	slog.Info("Processing movie request", "request_id", requestID, "tmdb_id", tmdbID)
	
	// Get Radarr instances to determine default settings
	radarrInstances, err := s.repo.GetArrServiceByType(ctx, "radarr")
	if err != nil {
		slog.Error("Failed to get Radarr instances", "error", err)
		return fmt.Errorf("failed to get Radarr instances: %w", err)
	}

	if len(radarrInstances) == 0 {
		slog.Error("No Radarr instances configured")
		return fmt.Errorf("no Radarr instances configured")
	}

	instance := radarrInstances[0]
	slog.Info("Using Radarr instance", 
		"name", instance.Name,
		"url", instance.BaseUrl,
		"root_folder", instance.RootFolderPath,
		"min_availability", instance.MinimumAvailability)

	// Add movie to Radarr with default settings
	slog.Info("Calling Radarr AddMovie", "tmdb_id", tmdbID)
	response, err := s.radarrService.AddMovie(
		ctx,
		tmdbID,
		1, // Default quality profile ID - could be made configurable
		instance.RootFolderPath,
		instance.MinimumAvailability,
	)
	if err != nil {
		slog.Error("Failed to add movie to Radarr", 
			"request_id", requestID, 
			"tmdb_id", tmdbID, 
			"error", err)
		return fmt.Errorf("failed to add movie to Radarr: %w", err)
	}

	slog.Info("Movie added to Radarr",
		"request_id", requestID,
		"tmdb_id", tmdbID,
		"radarr_id", response.ID,
		"title", response.Title)

	return nil
}

func (s *service) processSeriesRequest(ctx context.Context, requestID, tmdbID int64) error {
	// Get Sonarr instances to determine default settings
	sonarrInstances, err := s.repo.GetArrServiceByType(ctx, "sonarr")
	if err != nil {
		return fmt.Errorf("failed to get Sonarr instances: %w", err)
	}

	if len(sonarrInstances) == 0 {
		return fmt.Errorf("no Sonarr instances configured")
	}

	instance := sonarrInstances[0]

	// Add series to Sonarr with default settings
	response, err := s.sonarrService.AddSeries(
		ctx,
		tmdbID,
		1, // Default quality profile ID - could be made configurable
		instance.RootFolderPath,
	)
	if err != nil {
		slog.Error("Failed to add series to Sonarr", 
			"request_id", requestID, 
			"tmdb_id", tmdbID, 
			"error", err)
		return fmt.Errorf("failed to add series to Sonarr: %w", err)
	}

	slog.Info("Series added to Sonarr",
		"request_id", requestID,
		"tmdb_id", tmdbID,
		"sonarr_id", response.ID,
		"title", response.Title)

	return nil
}

// CheckRequestStatus checks if a request's media has been downloaded and fulfills it if needed
func (s *service) CheckRequestStatus(ctx context.Context, requestID int64) error {
	// Get the request details
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get request: %w", err)
	}

	// Only check approved requests that haven't been fulfilled yet
	if request.Status != "approved" {
		return nil
	}

	if !request.TmdbID.Valid {
		return fmt.Errorf("request %d has no TMDB ID", requestID)
	}

	tmdbID := request.TmdbID.Int64

	switch request.MediaType {
	case "movie":
		return s.checkMovieStatus(ctx, requestID, tmdbID)
	case "tv":
		return s.checkSeriesStatus(ctx, requestID, tmdbID)
	default:
		return fmt.Errorf("unsupported media type: %s", request.MediaType)
	}
}

func (s *service) checkMovieStatus(ctx context.Context, requestID, tmdbID int64) error {
	movie, err := s.radarrService.GetMovieByTMDBID(ctx, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to get movie status from Radarr: %w", err)
	}

	if movie == nil {
		// Movie not in Radarr yet
		return nil
	}

	// Check if the movie has been downloaded
	if movie.HasFile || movie.Downloaded {
		// Movie is downloaded, fulfill the request
		_, err := s.repo.FulfillRequest(ctx, requestID)
		if err != nil {
			return fmt.Errorf("failed to fulfill request: %w", err)
		}

		slog.Info("Request automatically fulfilled - movie downloaded",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"title", movie.Title)
	}

	return nil
}

func (s *service) checkSeriesStatus(ctx context.Context, requestID, tmdbID int64) error {
	series, err := s.sonarrService.GetSeriesByTMDBID(ctx, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to get series status from Sonarr: %w", err)
	}

	if series == nil {
		// Series not in Sonarr yet
		return nil
	}

	// Check if the series has downloaded episodes (at least some episodes)
	// You might want to adjust this logic based on your requirements
	if series.Statistics.EpisodeFileCount > 0 {
		// Series has downloaded episodes, fulfill the request
		_, err := s.repo.FulfillRequest(ctx, requestID)
		if err != nil {
			return fmt.Errorf("failed to fulfill request: %w", err)
		}

		slog.Info("Request automatically fulfilled - series has downloaded episodes",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"title", series.Title,
			"downloaded_episodes", series.Statistics.EpisodeFileCount)
	}

	return nil
}