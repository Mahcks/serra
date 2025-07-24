package request_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/integrations/emby"
	"github.com/mahcks/serra/internal/integrations/radarr"
	"github.com/mahcks/serra/internal/integrations/sonarr"
	"github.com/mahcks/serra/internal/services/season_availability"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type Service interface {
	ProcessApprovedRequest(ctx context.Context, requestID int64) error
	CheckRequestStatus(ctx context.Context, requestID int64) error
	CheckExistingAvailability(ctx context.Context, tmdbID int64, mediaType string, seasons []int) (*structures.ShowAvailability, error)
	RetryFailedRequests(ctx context.Context) error
}

type service struct {
	repo                      *repository.Queries
	radarrService             radarr.Service
	sonarrService             sonarr.Service
	seasonAvailabilityService *season_availability.SeasonAvailabilityService
	embyService               emby.Service
}

func New(repo *repository.Queries, radarrSvc radarr.Service, sonarrSvc sonarr.Service, integrations *integrations.Integration) Service {
	seasonSvc := season_availability.NewSeasonAvailabilityService(repo, integrations.Emby, integrations.TMDB)
	return &service{
		repo:                      repo,
		radarrService:             radarrSvc,
		sonarrService:             sonarrSvc,
		seasonAvailabilityService: seasonSvc,
		embyService:               integrations.Emby,
	}
}

// ProcessApprovedRequest automatically adds an approved request to Radarr/Sonarr
func (s *service) ProcessApprovedRequest(ctx context.Context, requestID int64) error {
	slog.Info("ProcessApprovedRequest called", "request_id", requestID)

	// First, mark the request as processing
	_, err := s.repo.UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
		Status: "processing",
		ID:     requestID,
	})
	if err != nil {
		slog.Error("Failed to update request status to processing", "request_id", requestID, "error", err)
		// Continue anyway - don't fail just because status update failed
	}

	// Get the request details
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		slog.Error("Failed to get request by ID", "request_id", requestID, "error", err)
		// Mark as failed before returning
		s.markRequestAsFailed(ctx, requestID, fmt.Sprintf("Failed to get request: %v", err))
		return fmt.Errorf("failed to get request: %w", err)
	}

	slog.Info("Processing request",
		"request_id", requestID,
		"title", request.Title,
		"media_type", request.MediaType,
		"status", request.Status,
		"tmdb_id", request.TmdbID)

	switch request.Status {
	case "approved", "processing", "failed":
		// Continue processing
		fmt.Println("Processing request with status:", request.Status)
	default:
		err := apiErrors.ErrRequestNotApproved().SetDetail("Current status: %s", request.Status)
		s.markRequestAsFailed(ctx, requestID, fmt.Sprintf("Invalid status for processing: %s", request.Status))
		return err
	}

	if !request.TmdbID.Valid {
		err := apiErrors.ErrMissingTMDBID()
		s.markRequestAsFailed(ctx, requestID, "Missing TMDB ID")
		return err
	}

	tmdbID := request.TmdbID.Int64

	var processErr error
	switch request.MediaType {
	case "movie":
		processErr = s.processMovieRequest(ctx, requestID, tmdbID)
	case "tv":
		processErr = s.processSeriesRequest(ctx, requestID, tmdbID)
	default:
		processErr = apiErrors.ErrInvalidMediaType().SetDetail("Unsupported media type: %s", request.MediaType)
	}

	// Handle processing result
	if processErr != nil {
		s.markRequestAsFailed(ctx, requestID, fmt.Sprintf("Processing failed: %v", processErr))
		return processErr
	}

	// Mark as successfully processed (back to approved for monitoring)
	_, err = s.repo.UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
		Status: "approved",
		ID:     requestID,
	})
	if err != nil {
		slog.Error("Failed to update request status back to approved", "request_id", requestID, "error", err)
	}

	slog.Info("Request processing completed successfully", "request_id", requestID)
	return nil
}

func (s *service) processMovieRequest(ctx context.Context, requestID, tmdbID int64) error {
	slog.Info("Processing movie request", "request_id", requestID, "tmdb_id", tmdbID)

	// TODO: For now, all requests are treated as non-4K since 4K detection is not implemented
	// In the future, this should get the request and check request.Is4K or similar field:
	// request, err := s.repo.GetRequestByID(ctx, requestID)
	// is4K := request.Is4K
	is4K := false

	// Get Radarr instances
	allRadarrInstances, err := s.repo.GetArrServiceByType(ctx, "radarr")
	if err != nil {
		slog.Error("Failed to get Radarr instances", "error", err)
		return fmt.Errorf("failed to get Radarr instances: %w", err)
	}

	if len(allRadarrInstances) == 0 {
		slog.Error("No Radarr instances configured")
		return apiErrors.ErrNoRadarrInstances()
	}

	// Filter instances by 4K preference
	var radarrInstances []repository.ArrService
	for _, instance := range allRadarrInstances {
		if instance.Is4k == is4K {
			radarrInstances = append(radarrInstances, instance)
		}
	}

	// Fallback to any instance if no matching type found
	if len(radarrInstances) == 0 {
		slog.Warn("No Radarr instances found for content type, using first available",
			"is_4k", is4K,
			"request_id", requestID)
		radarrInstances = allRadarrInstances
	}

	// Use first matching instance (could implement round-robin here later)
	instance := radarrInstances[0]
	slog.Info("Using Radarr instance",
		"name", instance.Name,
		"url", instance.BaseUrl,
		"root_folder", instance.RootFolderPath,
		"min_availability", instance.MinimumAvailability,
		"quality_profile", instance.QualityProfile,
		"is_4k", instance.Is4k,
		"content_type", map[bool]string{true: "4K", false: "regular"}[is4K])

	// Parse quality profile ID from stored string
	qualityProfileID, err := strconv.Atoi(instance.QualityProfile)
	if err != nil {
		slog.Error("Failed to parse quality profile ID",
			"quality_profile", instance.QualityProfile,
			"error", err)
		return apiErrors.ErrInvalidQualityProfile().SetDetail("Invalid quality profile '%s' for Radarr instance '%s'", instance.QualityProfile, instance.Name)
	}

	// Add movie to Radarr with configured quality profile
	slog.Info("Calling Radarr AddMovie",
		"tmdb_id", tmdbID,
		"quality_profile_id", qualityProfileID)
	response, err := s.radarrService.AddMovie(
		ctx,
		tmdbID,
		qualityProfileID,
		instance.RootFolderPath,
		instance.MinimumAvailability,
	)
	if err != nil {
		slog.Error("Failed to add movie to Radarr",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"error", err)
		return apiErrors.ErrRadarrConnection().SetDetail("Error: %s", err.Error())
	}

	slog.Info("Movie added to Radarr",
		"request_id", requestID,
		"tmdb_id", tmdbID,
		"radarr_id", response.ID,
		"title", response.Title)

	return nil
}

func (s *service) processSeriesRequest(ctx context.Context, requestID, tmdbID int64) error {
	slog.Info("Processing series request", "request_id", requestID, "tmdb_id", tmdbID)

	// Get the request to extract season information
	request, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		slog.Error("Failed to get request details", "request_id", requestID, "error", err)
		return fmt.Errorf("failed to get request details: %w", err)
	}

	// Parse seasons from request
	var seasons []int
	if request.Seasons.Valid && request.Seasons.String != "" {
		err := json.Unmarshal([]byte(request.Seasons.String), &seasons)
		if err != nil {
			slog.Error("Failed to parse seasons from request",
				"request_id", requestID,
				"seasons_json", request.Seasons.String,
				"error", err)
			return apiErrors.ErrSeasonParsingFailed()
		}
	}

	// TODO: For now, all requests are treated as non-4K since 4K detection is not implemented
	// In the future, this should check request.Is4K or similar field
	is4K := false

	// Get Sonarr instances
	allSonarrInstances, err := s.repo.GetArrServiceByType(ctx, "sonarr")
	if err != nil {
		slog.Error("Failed to get Sonarr instances", "error", err)
		return fmt.Errorf("failed to get Sonarr instances: %w", err)
	}

	if len(allSonarrInstances) == 0 {
		slog.Error("No Sonarr instances configured")
		return apiErrors.ErrNoSonarrInstances()
	}

	// Filter instances by 4K preference
	var sonarrInstances []repository.ArrService
	for _, instance := range allSonarrInstances {
		if instance.Is4k == is4K {
			sonarrInstances = append(sonarrInstances, instance)
		}
	}

	// Fallback to any instance if no matching type found
	if len(sonarrInstances) == 0 {
		slog.Warn("No Sonarr instances found for content type, using first available",
			"is_4k", is4K,
			"request_id", requestID)
		sonarrInstances = allSonarrInstances
	}

	// Use first matching instance (could implement round-robin here later)
	instance := sonarrInstances[0]
	slog.Info("Using Sonarr instance",
		"name", instance.Name,
		"url", instance.BaseUrl,
		"root_folder", instance.RootFolderPath,
		"quality_profile", instance.QualityProfile,
		"is_4k", instance.Is4k,
		"content_type", map[bool]string{true: "4K", false: "regular"}[is4K])

	// Parse quality profile ID from stored string
	qualityProfileID, err := strconv.Atoi(instance.QualityProfile)
	if err != nil {
		slog.Error("Failed to parse quality profile ID",
			"quality_profile", instance.QualityProfile,
			"error", err)
		return apiErrors.ErrInvalidQualityProfile().SetDetail("Invalid quality profile '%s' for Sonarr instance '%s'", instance.QualityProfile, instance.Name)
	}

	// Add series to Sonarr with configured quality profile and season-specific monitoring
	if len(seasons) > 0 {
		slog.Info("Calling Sonarr AddSeriesWithSeasons",
			"tmdb_id", tmdbID,
			"quality_profile_id", qualityProfileID,
			"seasons", seasons)
		response, err := s.sonarrService.AddSeriesWithSeasons(
			ctx,
			tmdbID,
			qualityProfileID,
			instance.RootFolderPath,
			seasons,
		)
		if err != nil {
			slog.Error("Failed to add series to Sonarr",
				"request_id", requestID,
				"tmdb_id", tmdbID,
				"error", err)
			return fmt.Errorf("failed to add series to Sonarr: %w", err)
		}

		slog.Info("Series added to Sonarr with specific seasons",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"sonarr_id", response.ID,
			"title", response.Title,
			"seasons", seasons)
	} else {
		slog.Info("Calling Sonarr AddSeries (all seasons)",
			"tmdb_id", tmdbID,
			"quality_profile_id", qualityProfileID)
		response, err := s.sonarrService.AddSeries(
			ctx,
			tmdbID,
			qualityProfileID,
			instance.RootFolderPath,
		)
		if err != nil {
			slog.Error("Failed to add series to Sonarr",
				"request_id", requestID,
				"tmdb_id", tmdbID,
				"error", err)
			return fmt.Errorf("failed to add series to Sonarr: %w", err)
		}

		slog.Info("Series added to Sonarr (all seasons)",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"sonarr_id", response.ID,
			"title", response.Title)
	}

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
	// First check if movie is downloaded in Radarr
	movie, err := s.radarrService.GetMovieByTMDBID(ctx, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to get movie status from Radarr: %w", err)
	}

	if movie == nil {
		// Movie not in Radarr yet
		return nil
	}

	// Check if the movie has been downloaded in Radarr
	if !(movie.HasFile || movie.Downloaded) {
		// Movie not downloaded yet in Radarr
		return nil
	}

	// Movie is downloaded in Radarr, now check if it's available in Emby/Jellyfin
	embyMovie, err := s.embyService.GetMovieByTMDBID(ctx, int(tmdbID))
	if err != nil {
		slog.Error("Failed to check movie availability in Emby",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"error", err)
		// Don't fail the entire process, just log the error
		return nil
	}

	if embyMovie == nil {
		// Movie downloaded in Radarr but not yet available in Emby
		slog.Debug("Movie downloaded in Radarr but not yet available in Emby",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"title", movie.Title)
		return nil
	}

	// Movie is both downloaded in Radarr AND available in Emby - fulfill the request
	_, err = s.repo.FulfillRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to fulfill request: %w", err)
	}

	slog.Info("Request automatically fulfilled - movie downloaded and available in media server",
		"request_id", requestID,
		"tmdb_id", tmdbID,
		"title", movie.Title)

	return nil
}

func (s *service) checkSeriesStatus(ctx context.Context, requestID, tmdbID int64) error {
	// First check if series has downloaded episodes in Sonarr
	series, err := s.sonarrService.GetSeriesByTMDBID(ctx, tmdbID)
	if err != nil {
		return fmt.Errorf("failed to get series status from Sonarr: %w", err)
	}

	if series == nil {
		// Series not in Sonarr yet
		return nil
	}

	// Check if the series has downloaded episodes (at least some episodes)
	if series.Statistics.EpisodeFileCount == 0 {
		// No episodes downloaded yet in Sonarr
		return nil
	}

	// Series has downloaded episodes in Sonarr, now check if it's available in Emby/Jellyfin
	embySeries, err := s.embyService.GetSeriesByTMDBID(ctx, int(tmdbID))
	if err != nil {
		slog.Error("Failed to check series availability in Emby",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"error", err)
		// Don't fail the entire process, just log the error
		return nil
	}

	if embySeries == nil {
		// Series has downloaded episodes in Sonarr but not yet available in Emby
		slog.Debug("Series has downloaded episodes in Sonarr but not yet available in Emby",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"title", series.Title,
			"downloaded_episodes", series.Statistics.EpisodeFileCount)
		return nil
	}

	// Verify that episodes are actually available in Emby by checking episode count
	episodes, err := s.embyService.GetEpisodesByTMDB(ctx, int(tmdbID))
	if err != nil {
		slog.Error("Failed to check episode availability in Emby",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"error", err)
		return nil
	}

	if len(episodes) == 0 {
		// Series exists in Emby but no episodes available yet
		slog.Debug("Series exists in Emby but no episodes available yet",
			"request_id", requestID,
			"tmdb_id", tmdbID,
			"title", series.Title)
		return nil
	}

	// Series has both downloaded episodes in Sonarr AND available episodes in Emby - fulfill the request
	_, err = s.repo.FulfillRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to fulfill request: %w", err)
	}

	slog.Info("Request automatically fulfilled - series has downloaded episodes and available in media server",
		"request_id", requestID,
		"tmdb_id", tmdbID,
		"title", series.Title,
		"downloaded_episodes", series.Statistics.EpisodeFileCount,
		"available_episodes", len(episodes))

	return nil
}

// CheckExistingAvailability checks what seasons/episodes are already available for a given TMDB ID
func (s *service) CheckExistingAvailability(ctx context.Context, tmdbID int64, mediaType string, seasons []int) (*structures.ShowAvailability, error) {
	if mediaType == "movie" {
		// For movies, we don't need season-level checking
		// TODO: Check if movie exists in media server
		return &structures.ShowAvailability{
			TmdbID:        int(tmdbID),
			OverallStatus: "not_available", // Simplified for now
		}, nil
	}

	if mediaType == "tv" {
		// Sync current availability from media server
		err := s.seasonAvailabilityService.SyncShowAvailability(ctx, int(tmdbID))
		if err != nil {
			slog.Error("Failed to sync show availability", "tmdb_id", tmdbID, "error", err)
			// Continue with existing data if sync fails
		}

		// Get current availability
		availability, err := s.seasonAvailabilityService.GetSeasonAvailability(ctx, int(tmdbID))
		if err != nil {
			slog.Error("Failed to get season availability", "tmdb_id", tmdbID, "error", err)
			return &structures.ShowAvailability{
				TmdbID:        int(tmdbID),
				OverallStatus: "not_available",
				Seasons:       []structures.SeasonAvailability{},
			}, nil
		}

		return availability, nil
	}

	return nil, fmt.Errorf("unsupported media type: %s", mediaType)
}

// markRequestAsFailed marks a request as failed with an error message
func (s *service) markRequestAsFailed(ctx context.Context, requestID int64, errorMessage string) {
	_, err := s.repo.UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
		Status: "failed",
		ID:     requestID,
	})
	if err != nil {
		slog.Error("Failed to update request status to failed", "request_id", requestID, "error", err)
	}
	
	slog.Error("Request marked as failed", "request_id", requestID, "reason", errorMessage)
}

// RetryFailedRequests attempts to retry all failed requests
func (s *service) RetryFailedRequests(ctx context.Context) error {
	failedRequests, err := s.repo.GetRequestsByStatus(ctx, "failed")
	if err != nil {
		return fmt.Errorf("failed to get failed requests: %w", err)
	}

	slog.Info("Retrying failed requests", "count", len(failedRequests))

	for _, request := range failedRequests {
		slog.Info("Retrying failed request", "request_id", request.ID, "title", request.Title)
		
		// Reset status to approved to trigger processing
		_, err := s.repo.UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
			Status: "approved",
			ID:     request.ID,
		})
		if err != nil {
			slog.Error("Failed to reset failed request status", "request_id", request.ID, "error", err)
			continue
		}

		// Process the request
		if err := s.ProcessApprovedRequest(ctx, request.ID); err != nil {
			slog.Error("Failed to retry request", "request_id", request.ID, "error", err)
			// markRequestAsFailed is already called in ProcessApprovedRequest
		} else {
			slog.Info("Successfully retried request", "request_id", request.ID)
		}
	}

	return nil
}
