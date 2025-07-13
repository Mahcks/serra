package requests

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// checkUserPermission checks if a user has a specific permission
func (rg *RouteGroup) checkUserPermissionForDelete(ctx context.Context, userID, permission string) (bool, error) {
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

func (rg *RouteGroup) DeleteRequest(ctx *respond.Ctx) error {
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

	// Get the existing request to check ownership
	existingRequest, err := rg.gctx.Crate().Sqlite.Query().GetRequestByID(ctx.Context(), requestID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("Request not found")
		}
		slog.Error("Failed to get request by ID", "error", err, "request_id", requestID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve request")
	}

	// Check permissions - users can delete their own requests, admins can delete any
	canManage := user.IsAdmin
	if !user.IsAdmin {
		var err error
		canManage, err = rg.checkUserPermissionForDelete(ctx.Context(), user.ID, permissions.RequestsManage)
		if err != nil {
			slog.Error("Failed to check manage permission", "error", err)
			return apiErrors.ErrInternalServerError().SetDetail("Permission check failed")
		}
	}
	isOwner := existingRequest.UserID == user.ID

	if !canManage && !isOwner {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to delete this request")
	}

	// Don't allow deletion of fulfilled requests unless user is admin
	if existingRequest.Status == "fulfilled" && !canManage {
		return apiErrors.ErrBadRequest().SetDetail("Cannot delete fulfilled requests")
	}

	// Delete the request
	err = rg.gctx.Crate().Sqlite.Query().DeleteRequest(ctx.Context(), requestID)
	if err != nil {
		slog.Error("Failed to delete request", "error", err, "request_id", requestID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to delete request")
	}

	slog.Info("Request deleted", 
		"request_id", requestID, 
		"deleted_by", user.ID, 
		"original_user_id", existingRequest.UserID,
		"was_admin_delete", canManage)

	return ctx.JSON(map[string]interface{}{
		"message": "Request deleted successfully",
		"id":      requestID,
	})
}