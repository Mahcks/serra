package requests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)


// checkUserPermissionCreate checks if a user has a specific permission
func (rg *RouteGroup) checkUserPermissionCreate(ctx context.Context, userID, permission string) (bool, error) {
	userPermissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if the user has owner permission (grants all access)
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permissions.Owner {
			return true, nil
		}
	}

	// Check if the permission exists in the user's permissions
	for _, userPerm := range userPermissions {
		if userPerm.PermissionID == permission {
			return true, nil
		}
	}

	return false, nil
}

func (rg *RouteGroup) CreateRequest(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Parse request body
	var req structures.CreateRequestRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Check permissions
	var requiredPermission string
	if req.MediaType == "movie" {
		requiredPermission = permissions.RequestMovies
	} else if req.MediaType == "tv" {
		requiredPermission = permissions.RequestSeries
	} else {
		return apiErrors.ErrInvalidMediaType().SetDetail("Media type '%s' is not supported", req.MediaType)
	}

	hasRequestPermission := user.IsAdmin
	if !hasRequestPermission {
		var err error
		hasRequestPermission, err = rg.checkUserPermissionCreate(ctx.Context(), user.ID, requiredPermission)
		if err != nil {
			slog.Error("Failed to check permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
	}

	if !hasRequestPermission {
		var mediaTypeFriendly string
		if req.MediaType == "movie" {
			mediaTypeFriendly = "movies"
		} else {
			mediaTypeFriendly = "TV shows"
		}
		return apiErrors.ErrNoRequestPermission().SetDetail("You need permission to request %s", mediaTypeFriendly)
	}

	// Check for duplicate requests with sophisticated season handling
	if err := rg.checkForDuplicateRequest(ctx.Context(), user.ID, req.MediaType, req.TmdbID, req.Seasons); err != nil {
		return err
	}

	// Validate on_behalf_of user exists if provided
	if req.OnBehalfOf != nil && *req.OnBehalfOf != "" {
		// Check if user has permission to create requests on behalf of others
		canManageRequests := user.IsAdmin
		if !canManageRequests {
			var err error
			canManageRequests, err = rg.checkUserPermissionCreate(ctx.Context(), user.ID, permissions.RequestsManage)
			if err != nil {
				slog.Error("Failed to check manage permission", "error", err)
				return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
			}
		}
		if !canManageRequests {
			return apiErrors.ErrNoManagePermission().SetDetail("You cannot create requests on behalf of other users")
		}

		// Verify the target user exists
		_, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), *req.OnBehalfOf)
		if err != nil {
			if err == sql.ErrNoRows {
				return apiErrors.ErrBadRequest().SetDetail("User specified in on_behalf_of does not exist")
			}
			slog.Error("Failed to verify on_behalf_of user", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to verify user")
		}
	}

	// Create the request
	params := repository.CreateRequestParams{
		UserID:    user.ID,
		MediaType: req.MediaType,
		TmdbID:    sql.NullInt64{Int64: req.TmdbID, Valid: true},
		Title:     sql.NullString{String: req.Title, Valid: true},
		Status:    "", // Will be set below based on auto-approval permission
		Notes:     sql.NullString{},
		PosterUrl: sql.NullString{},
		OnBehalfOf: sql.NullString{},
		Seasons:   sql.NullString{},
		SeasonStatuses: sql.NullString{},
	}

	if req.Notes != nil {
		params.Notes = sql.NullString{String: *req.Notes, Valid: true}
	}

	if req.PosterURL != nil {
		params.PosterUrl = sql.NullString{String: *req.PosterURL, Valid: true}
	}

	if req.OnBehalfOf != nil {
		params.OnBehalfOf = sql.NullString{String: *req.OnBehalfOf, Valid: true}
	}

	// Handle seasons for TV shows
	if req.MediaType == "tv" && len(req.Seasons) > 0 {
		// Check existing availability before creating request
		availability, err := rg.requestProcessor.CheckExistingAvailability(ctx.Context(), req.TmdbID, req.MediaType, req.Seasons)
		if err != nil {
			slog.Error("Failed to check existing availability", "error", err)
			// Continue with request creation if availability check fails
		}

		seasonsJSON, err := json.Marshal(req.Seasons)
		if err != nil {
			slog.Error("Failed to marshal seasons", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to process seasons")
		}
		params.Seasons = sql.NullString{String: string(seasonsJSON), Valid: true}

		// Initialize season statuses based on existing availability
		seasonStatuses := make(map[string]structures.SeasonInfo)
		for _, season := range req.Seasons {
			seasonKey := fmt.Sprintf("%d", season)
			
			// Check if this season is already available
			var seasonInfo structures.SeasonInfo
			isAvailable := false
			if availability != nil {
				for _, availSeason := range availability.Seasons {
					if availSeason.SeasonNumber == season && availSeason.IsComplete {
						isAvailable = true
						seasonInfo = structures.SeasonInfo{
							Status:           "fulfilled",
							Episodes:         fmt.Sprintf("%d/%d", availSeason.AvailableEpisodes, availSeason.EpisodeCount),
							AvailableEpisodes: availSeason.AvailableEpisodes,
							TotalEpisodes:     availSeason.EpisodeCount,
							LastUpdated:       availSeason.LastUpdated,
						}
						break
					}
				}
			}
			
			// If not available, set as pending
			if !isAvailable {
				seasonInfo = structures.SeasonInfo{
					Status:           "pending",
					Episodes:         "0/0", // Will be updated when TMDB data is fetched
					AvailableEpisodes: 0,
					TotalEpisodes:     0,
					LastUpdated:       "", // Will be set by database
				}
			}
			
			seasonStatuses[seasonKey] = seasonInfo
		}
		
		statusJSON, err := json.Marshal(seasonStatuses)
		if err != nil {
			slog.Error("Failed to marshal season statuses", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to process season statuses")
		}
		params.SeasonStatuses = sql.NullString{String: string(statusJSON), Valid: true}
	}

	// Check if user has auto-approval permission for this specific media type
	// TODO: In the future, add support for 4K detection from frontend
	// For now, treat all requests as regular (non-4K) content
	var autoApprovalPermission string
	if req.MediaType == "movie" {
		autoApprovalPermission = permissions.RequestAutoApproveMovies
	} else if req.MediaType == "tv" {
		autoApprovalPermission = permissions.RequestAutoApproveSeries
	}

	hasAutoApproval := user.IsAdmin
	if !hasAutoApproval && autoApprovalPermission != "" {
		var err error
		hasAutoApproval, err = rg.checkUserPermissionCreate(ctx.Context(), user.ID, autoApprovalPermission)
		if err != nil {
			slog.Error("Failed to check auto-approval permission", "error", err, "permission", autoApprovalPermission)
			// Continue with normal flow if permission check fails
		}
	}

	// Set status based on auto-approval permission
	if hasAutoApproval {
		params.Status = "approved"
		slog.Info("Request auto-approved", "user_id", user.ID, "media_type", req.MediaType, "tmdb_id", req.TmdbID, "permission", autoApprovalPermission)
	} else {
		params.Status = "pending"
	}

	request, err := rg.gctx.Crate().Sqlite.Query().CreateRequest(ctx.Context(), params)
	if err != nil {
		utils.LogErrorWithStack("Failed to create request", err, 
			"user_id", user.ID, 
			"media_type", req.MediaType, 
			"tmdb_id", req.TmdbID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create request")
	}

	slog.Info("Request created", 
		"request_id", request.ID, 
		"user_id", user.ID, 
		"media_type", req.MediaType, 
		"tmdb_id", req.TmdbID,
		"title", req.Title,
		"status", params.Status,
		"auto_approved", hasAutoApproval)

	// If request was auto-approved, automatically process it with proper error handling
	if hasAutoApproval {
		slog.Info("Auto-approved request - triggering automation", 
			"request_id", request.ID, 
			"title", request.Title,
			"media_type", request.MediaType,
			"tmdb_id", request.TmdbID)
		
		// Process the request asynchronously with proper error handling and status updates
		title := request.Title.String
		if !request.Title.Valid {
			title = "Unknown Title"
		}
		go rg.processAutoApprovedRequestWithRecovery(request.ID, title)
	}

	// Convert repository.Request to structures.Request for proper JSON response
	apiRequest := structures.Request{
		ID:        request.ID,
		UserID:    request.UserID,
		MediaType: request.MediaType,
		Status:    request.Status,
		CreatedAt: request.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: request.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	
	if request.TmdbID.Valid {
		tmdbID := request.TmdbID.Int64
		apiRequest.TmdbID = &tmdbID
	}
	if request.Title.Valid {
		apiRequest.Title = request.Title.String
	}
	if request.Notes.Valid {
		apiRequest.Notes = request.Notes.String
	}
	if request.FulfilledAt.Valid {
		fulfilledAt := request.FulfilledAt.Time.Format("2006-01-02T15:04:05Z")
		apiRequest.FulfilledAt = fulfilledAt
	}
	if request.ApproverID.Valid {
		apiRequest.ApproverID = request.ApproverID.String
	}
	if request.OnBehalfOf.Valid {
		apiRequest.OnBehalfOf = request.OnBehalfOf.String
	}
	if request.PosterUrl.Valid {
		apiRequest.PosterURL = request.PosterUrl.String
	}
	if request.Seasons.Valid {
		var seasons []int
		if err := json.Unmarshal([]byte(request.Seasons.String), &seasons); err == nil {
			apiRequest.Seasons = seasons
		}
	}
	if request.SeasonStatuses.Valid {
		var seasonStatuses map[string]structures.SeasonInfo
		if err := json.Unmarshal([]byte(request.SeasonStatuses.String), &seasonStatuses); err == nil {
			apiRequest.SeasonStatuses = seasonStatuses
		}
	}

	return ctx.JSON(apiRequest)
}

// checkForDuplicateRequest checks for duplicate requests with sophisticated season handling
func (rg *RouteGroup) checkForDuplicateRequest(ctx context.Context, userID, mediaType string, tmdbID int64, requestedSeasons []int) error {
	var seasonsJSON sql.NullString
	if len(requestedSeasons) > 0 {
		seasonsBytes, err := json.Marshal(requestedSeasons)
		if err != nil {
			slog.Error("Failed to marshal seasons", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to process season data")
		}
		seasonsJSON = sql.NullString{String: string(seasonsBytes), Valid: true}
	}

	// For movies, check exact duplicate (including null seasons)
	if mediaType == "movie" {
		existingRequest, err := rg.gctx.Crate().Sqlite.Query().CheckExistingRequest(ctx, repository.CheckExistingRequestParams{
			MediaType: mediaType,
			TmdbID:    sql.NullInt64{Int64: tmdbID, Valid: true},
			UserID:    userID,
			Seasons:   seasonsJSON,
		})

		if err == nil {
			// Exact duplicate found
			return apiErrors.ErrConflict().SetDetail(fmt.Sprintf("You already have a %s request for this movie", existingRequest.Status))
		} else if err != sql.ErrNoRows {
			// Database error
			slog.Error("Failed to check existing movie request", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to check existing request")
		}
		return nil // No duplicate found
	}

	// For TV shows, we need more sophisticated checking
	if mediaType == "tv" {
		// First, check for exact duplicate (same seasons)
		existingRequest, err := rg.gctx.Crate().Sqlite.Query().CheckExistingRequest(ctx, repository.CheckExistingRequestParams{
			MediaType: mediaType,
			TmdbID:    sql.NullInt64{Int64: tmdbID, Valid: true},
			UserID:    userID,
			Seasons:   seasonsJSON,
		})

		if err == nil {
			// Exact duplicate found
			return apiErrors.ErrConflict().SetDetail(fmt.Sprintf("You already have a %s request for these exact seasons", existingRequest.Status))
		} else if err != sql.ErrNoRows {
			// Database error
			slog.Error("Failed to check existing TV request", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to check existing request")
		}

		// Check for conflicting requests (overlapping seasons or whole series requests)
		existingRequests, err := rg.gctx.Crate().Sqlite.Query().CheckExistingRequestAnySeasons(ctx, repository.CheckExistingRequestAnySeasonsParams{
			MediaType: mediaType,
			TmdbID:    sql.NullInt64{Int64: tmdbID, Valid: true},
			UserID:    userID,
		})

		if err != nil && err != sql.ErrNoRows {
			slog.Error("Failed to check existing TV requests", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to check existing requests")
		}

		// Analyze conflicts
		for _, existing := range existingRequests {
			conflict, message := rg.checkSeasonConflict(requestedSeasons, existing.Seasons.String, existing.Status)
			if conflict {
				return apiErrors.ErrConflict().SetDetail(message)
			}
		}

		return nil // No conflicts found
	}

	return nil
}

// checkSeasonConflict checks if the requested seasons conflict with an existing request
func (rg *RouteGroup) checkSeasonConflict(requestedSeasons []int, existingSeasonsJSON string, existingStatus string) (bool, string) {
	// If no seasons requested, this is a whole series request
	if len(requestedSeasons) == 0 {
		if existingSeasonsJSON == "" {
			// Both are whole series requests
			return true, fmt.Sprintf("You already have a %s request for the entire series", existingStatus)
		}
		// Requesting whole series but existing is season-specific
		return true, fmt.Sprintf("You already have a %s request for specific seasons. Cannot request entire series", existingStatus)
	}

	// If existing request has no seasons, it's a whole series request
	if existingSeasonsJSON == "" {
		return true, fmt.Sprintf("You already have a %s request for the entire series", existingStatus)
	}

	// Both are season-specific, check for overlaps
	var existingSeasons []int
	if err := json.Unmarshal([]byte(existingSeasonsJSON), &existingSeasons); err != nil {
		slog.Error("Failed to unmarshal existing seasons", "error", err)
		// If we can't parse, assume conflict to be safe
		return true, "Cannot determine season conflict due to data error"
	}

	// Check for overlapping seasons
	requestedSet := make(map[int]bool)
	for _, season := range requestedSeasons {
		requestedSet[season] = true
	}

	var overlapping []int
	for _, season := range existingSeasons {
		if requestedSet[season] {
			overlapping = append(overlapping, season)
		}
	}

	if len(overlapping) > 0 {
		return true, fmt.Sprintf("You already have a %s request for season(s) %v", existingStatus, overlapping)
	}

	return false, ""
}

// processAutoApprovedRequestWithRecovery handles auto-approved request processing with proper error handling
func (rg *RouteGroup) processAutoApprovedRequestWithRecovery(requestID int64, title string) {
	// Add panic recovery to prevent crashes
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Panic during auto-approved request processing",
				"request_id", requestID,
				"title", title,
				"panic", r)
			
			// Mark request as failed
			rg.markRequestAsFailed(requestID, fmt.Sprintf("Processing panic: %v", r))
		}
	}()

	// Create a context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*60*time.Second) // 5 minute timeout
	defer cancel()

	slog.Info("Starting auto-approved request processing",
		"request_id", requestID,
		"title", title)

	// Update status to "processing" to indicate work is in progress
	_, err := rg.gctx.Crate().Sqlite.Query().UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
		Status: "processing",
		ID:     requestID,
	})
	if err != nil {
		slog.Error("Failed to update request status to processing",
			"request_id", requestID,
			"error", err)
		// Continue with processing, but this is concerning
	}

	// Attempt to process the request
	err = rg.processApprovedRequest(ctx, requestID)
	if err != nil {
		slog.Error("Failed to process auto-approved request",
			"request_id", requestID,
			"title", title,
			"error", err)
		
		// Mark request as failed with specific error message
		rg.markRequestAsFailed(requestID, err.Error())
		return
	}

	slog.Info("Successfully processed auto-approved request",
		"request_id", requestID,
		"title", title)
}

// markRequestAsFailed updates a request status to "failed" with error information
func (rg *RouteGroup) markRequestAsFailed(requestID int64, errorMessage string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Update status to "failed"
	_, err := rg.gctx.Crate().Sqlite.Query().UpdateRequestStatusOnly(ctx, repository.UpdateRequestStatusOnlyParams{
		Status: "failed",
		ID:     requestID,
	})
	if err != nil {
		slog.Error("Failed to mark request as failed",
			"request_id", requestID,
			"error", err,
			"original_error", errorMessage)
		return
	}

	slog.Info("Marked request as failed",
		"request_id", requestID,
		"error_message", errorMessage)
}