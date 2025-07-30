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


func (rg *RouteGroup) UpdateRequest(ctx *respond.Ctx) error {
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

	// Parse request body
	var req structures.UpdateRequestRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Get the existing request
	existingRequest, err := rg.gctx.Crate().Sqlite.Query().GetRequestByID(ctx.Context(), requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Request not found")
		}
		slog.Error("Failed to get request by ID", "error", err, "request_id", requestID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve request")
	}

	// Check permissions
	canApprove := user.IsAdmin
	canManage := user.IsAdmin
	if !user.IsAdmin {
		var err error
		canApprove, err = rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsApprove)
		if err != nil {
			slog.Error("Failed to check approve permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
		canManage, err = rg.checkUserPermission(ctx.Context(), user.ID, permissions.RequestsManage)
		if err != nil {
			slog.Error("Failed to check manage permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
	}
	isOwner := existingRequest.UserID == user.ID

	// Only owners can cancel their own requests (set to "denied")
	// Only admins can approve/deny/fulfill requests
	if req.Status == "denied" && isOwner {
		// User canceling their own request - this is allowed
	} else if req.Status == "approved" || req.Status == "fulfilled" {
		if !canApprove && !canManage {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to approve or fulfill requests")
		}
	} else if req.Status == "denied" && !isOwner {
		if !canApprove && !canManage {
			return apiErrors.ErrForbidden().SetDetail("You don't have permission to deny requests")
		}
	} else {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to update this request")
	}

	// Special handling for fulfilled status
	if req.Status == "fulfilled" {
		updatedRequest, err := rg.gctx.Crate().Sqlite.Query().FulfillRequest(ctx.Context(), requestID)
		if err != nil {
			slog.Error("Failed to fulfill request", "error", err, "request_id", requestID)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to fulfill request")
		}

		slog.Info("Request fulfilled", 
			"request_id", requestID, 
			"approver_id", user.ID, 
			"original_user_id", existingRequest.UserID)

		// Notify user that their media is now available
		if existingRequest.Title.Valid {
			go func() {
				bgCtx := context.Background()
				var tmdbID *int64
				if existingRequest.TmdbID.Valid {
					tmdbID = &existingRequest.TmdbID.Int64
				}
				
				err := rg.gctx.Crate().NotificationService.NotifyMediaAvailable(
					bgCtx,
					existingRequest.UserID,
					existingRequest.Title.String,
					existingRequest.MediaType,
					tmdbID,
				)
				if err != nil {
					slog.Error("Failed to send media available notification", "error", err, "request_id", requestID)
				}
			}()
		}

		return ctx.JSON(updatedRequest)
	}

	// Update request status
	var approverID sql.NullString
	if req.Status != "pending" {
		approverID = sql.NullString{String: user.ID, Valid: true}
	}

	updatedRequest, err := rg.gctx.Crate().Sqlite.Query().UpdateRequestStatus(ctx.Context(), repository.UpdateRequestStatusParams{
		Status:     req.Status,
		ApproverID: approverID,
		ID:         requestID,
	})

	if err != nil {
		slog.Error("Failed to update request status", "error", err, "request_id", requestID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to update request")
	}

	slog.Info("Request status updated", 
		"request_id", requestID, 
		"new_status", req.Status, 
		"approver_id", user.ID, 
		"original_user_id", existingRequest.UserID)

	// Send notifications for status changes
	if existingRequest.Title.Valid && req.Status != "pending" {
		go func() {
			bgCtx := context.Background()
			var tmdbID *int64
			if existingRequest.TmdbID.Valid {
				tmdbID = &existingRequest.TmdbID.Int64
			}
			requestIDStr := strconv.FormatInt(requestID, 10)
			
			switch req.Status {
			case "approved":
				err := rg.gctx.Crate().NotificationService.NotifyRequestApproved(
					bgCtx,
					existingRequest.UserID,
					existingRequest.Title.String,
					existingRequest.MediaType,
					tmdbID,
					&requestIDStr,
				)
				if err != nil {
					slog.Error("Failed to send request approved notification", "error", err, "request_id", requestID)
				}
			case "denied":
				reason := ""
				if req.Notes != nil {
					reason = *req.Notes
				}
				err := rg.gctx.Crate().NotificationService.NotifyRequestDenied(
					bgCtx,
					existingRequest.UserID,
					existingRequest.Title.String,
					existingRequest.MediaType,
					reason,
					tmdbID,
					&requestIDStr,
				)
				if err != nil {
					slog.Error("Failed to send request denied notification", "error", err, "request_id", requestID)
				}
			}
		}()
	}

	// If request was approved, automatically process it
	if req.Status == "approved" {
		slog.Info("Request approved - triggering automation", 
			"request_id", requestID, 
			"title", existingRequest.Title,
			"media_type", existingRequest.MediaType,
			"tmdb_id", existingRequest.TmdbID)
		
		go func() {
			// Use background context to avoid cancellation issues
			bgCtx := context.Background()
			if err := rg.processApprovedRequest(bgCtx, requestID); err != nil {
				slog.Error("Failed to process approved request - request marked as failed",
					"request_id", requestID,
					"title", existingRequest.Title,
					"error", err)
				// Status is already updated to failed in the processor
			} else {
				slog.Info("Successfully processed approved request",
					"request_id", requestID,
					"title", existingRequest.Title)
			}
		}()
	}

	return ctx.JSON(updatedRequest)
}