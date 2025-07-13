package requests

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

// checkUserPermission checks if a user has a specific permission
func (rg *RouteGroup) checkUserPermission(ctx context.Context, userID, permission string) (bool, error) {
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

// GetAllRequests returns all requests (admin only)
func (rg *RouteGroup) GetAllRequests(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check admin permissions - admins or users with specific permission
	if !user.IsAdmin {
		hasPermission, err := rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsView)
		if err != nil {
			slog.Error("Failed to check permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		if !hasPermission {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to view all requests")
		}
	}

	requests, err := rg.gctx.Crate().Sqlite.Query().GetAllRequests(ctx.Context())
	if err != nil {
		slog.Error("Failed to get all requests", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve requests")
	}

	// Convert repository.Request to structures.Request and fetch usernames
	var apiRequests []structures.Request
	for _, req := range requests {
		// Get username for the user_id
		user, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), req.UserID)
		username := req.UserID // fallback to user ID if username fetch fails
		if err == nil {
			username = user.Username
		}

		apiRequest := structures.Request{
			ID:        req.ID,
			UserID:    req.UserID,
			Username:  username,
			MediaType: req.MediaType,
			Status:    req.Status,
			CreatedAt: req.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: req.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		
		if req.TmdbID.Valid {
			tmdbID := req.TmdbID.Int64
			apiRequest.TmdbID = &tmdbID
		}
		if req.Title.Valid {
			apiRequest.Title = req.Title.String
		}
		if req.Notes.Valid {
			apiRequest.Notes = req.Notes.String
		}
		if req.FulfilledAt.Valid {
			fulfilledAt := req.FulfilledAt.Time.Format("2006-01-02T15:04:05Z")
			apiRequest.FulfilledAt = fulfilledAt
		}
		if req.ApproverID.Valid {
			apiRequest.ApproverID = req.ApproverID.String
		}
		if req.OnBehalfOf.Valid {
			apiRequest.OnBehalfOf = req.OnBehalfOf.String
		}
		if req.PosterUrl.Valid {
			apiRequest.PosterURL = req.PosterUrl.String
		}
		
		apiRequests = append(apiRequests, apiRequest)
	}

	return ctx.JSON(apiRequests)
}

// GetUserRequests returns current user's requests
func (rg *RouteGroup) GetUserRequests(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Get all requests and filter by user ID in the application
	allRequests, err := rg.gctx.Crate().Sqlite.Query().GetAllRequests(ctx.Context())
	if err != nil {
		slog.Error("Failed to get all requests", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve requests")
	}

	// Filter requests for the current user
	var requests []repository.Request
	for _, req := range allRequests {
		if req.UserID == user.ID || (req.OnBehalfOf.Valid && req.OnBehalfOf.String == user.ID) {
			requests = append(requests, req)
		}
	}

	// Convert repository.Request to structures.Request
	var apiRequests []structures.Request
	for _, req := range requests {
		apiRequest := structures.Request{
			ID:        req.ID,
			UserID:    req.UserID,
			MediaType: req.MediaType,
			Status:    req.Status,
			CreatedAt: req.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: req.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		
		if req.TmdbID.Valid {
			tmdbID := req.TmdbID.Int64
			apiRequest.TmdbID = &tmdbID
		}
		if req.Title.Valid {
			apiRequest.Title = req.Title.String
		}
		if req.Notes.Valid {
			apiRequest.Notes = req.Notes.String
		}
		if req.FulfilledAt.Valid {
			fulfilledAt := req.FulfilledAt.Time.Format("2006-01-02T15:04:05Z")
			apiRequest.FulfilledAt = fulfilledAt
		}
		if req.ApproverID.Valid {
			apiRequest.ApproverID = req.ApproverID.String
		}
		if req.OnBehalfOf.Valid {
			apiRequest.OnBehalfOf = req.OnBehalfOf.String
		}
		if req.PosterUrl.Valid {
			apiRequest.PosterURL = req.PosterUrl.String
		}
		
		apiRequests = append(apiRequests, apiRequest)
	}

	return ctx.JSON(apiRequests)
}

// GetPendingRequests returns pending requests (admin only)
func (rg *RouteGroup) GetPendingRequests(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check admin permissions
	if !user.IsAdmin {
		hasPermission, err := rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsView)
		if err != nil {
			slog.Error("Failed to check permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		if !hasPermission {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to view pending requests")
		}
	}

	requests, err := rg.gctx.Crate().Sqlite.Query().GetPendingRequests(ctx.Context())
	if err != nil {
		slog.Error("Failed to get pending requests", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve pending requests")
	}

	return ctx.JSON(requests)
}

// GetRequestByID returns specific request by ID
func (rg *RouteGroup) GetRequestByID(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Parse request ID
	requestIDStr := ctx.Params("id")
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request ID")
	}

	// Get the request
	request, err := rg.gctx.Crate().Sqlite.Query().GetRequestByID(ctx.Context(), requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Request not found")
		}
		slog.Error("Failed to get request by ID", "error", err, "request_id", requestID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve request")
	}

	// Check permissions - users can only see their own requests unless they're admin
	canViewAll := user.IsAdmin
	if !canViewAll {
		hasPermission, err := rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsView)
		if err != nil {
			slog.Error("Failed to check permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		canViewAll = hasPermission
	}

	isOwner := request.UserID == user.ID
	isOnBehalfOf := request.OnBehalfOf.Valid && request.OnBehalfOf.String == user.ID

	if !canViewAll && !isOwner && !isOnBehalfOf {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to view this request")
	}

	return ctx.JSON(request)
}

// GetRequestStatistics returns request statistics (admin only)
func (rg *RouteGroup) GetRequestStatistics(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check admin permissions
	if !user.IsAdmin {
		hasPermission, err := rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsView)
		if err != nil {
			slog.Error("Failed to check permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		if !hasPermission {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to view request statistics")
		}
	}

	stats, err := rg.gctx.Crate().Sqlite.Query().GetRequestStatistics(ctx.Context())
	if err != nil {
		slog.Error("Failed to get request statistics", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve request statistics")
	}

	// Convert to structures.RequestStatistics
	apiStats := structures.RequestStatistics{
		TotalRequests:     stats.TotalRequests,
		PendingRequests:   stats.PendingRequests,
		ApprovedRequests:  stats.ApprovedRequests,
		DeniedRequests:    stats.DeniedRequests,
		FulfilledRequests: stats.FulfilledRequests,
	}

	return ctx.JSON(apiStats)
}