package requests

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
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
		return apiErrors.ErrBadRequest().SetDetail("Invalid media type")
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
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to create requests for this media type")
	}

	// Check if request already exists for this user and media
	existingRequest, err := rg.gctx.Crate().Sqlite.Query().CheckExistingRequest(ctx.Context(), repository.CheckExistingRequestParams{
		MediaType: req.MediaType,
		TmdbID:    sql.NullInt64{Int64: req.TmdbID, Valid: true},
		UserID:    user.ID,
	})

	if err == nil {
		// Request already exists
		return apiErrors.ErrConflict().SetDetail(fmt.Sprintf("You already have a %s request for this media", existingRequest.Status))
	} else if err != sql.ErrNoRows {
		// Database error
		slog.Error("Failed to check existing request", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to check existing request")
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
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to create requests on behalf of others")
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
		slog.Error("Failed to create request", "error", err)
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

	// If request was auto-approved, automatically process it
	if hasAutoApproval {
		slog.Info("Auto-approved request - triggering automation", 
			"request_id", request.ID, 
			"title", request.Title,
			"media_type", request.MediaType,
			"tmdb_id", request.TmdbID)
		
		go func() {
			// Use background context to avoid cancellation issues
			bgCtx := context.Background()
			if err := rg.processApprovedRequest(bgCtx, request.ID); err != nil {
				slog.Error("Failed to process auto-approved request",
					"request_id", request.ID,
					"title", request.Title,
					"error", err)
			} else {
				slog.Info("Successfully triggered automation for auto-approved request",
					"request_id", request.ID,
					"title", request.Title)
			}
		}()
	}

	return ctx.JSON(request)
}