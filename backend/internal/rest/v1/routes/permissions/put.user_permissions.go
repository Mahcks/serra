package permissions

import (
	"database/sql"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// BulkUpdateUserPermissions updates all permissions for a user at once
func (rg *RouteGroup) BulkUpdateUserPermissions(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	var req structures.BulkUpdatePermissionsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Handle null permissions from frontend
	if req.Permissions == nil {
		req.Permissions = []string{}
	}

	slog.Info("BulkUpdateUserPermissions request", "user_id", userID, "requested_permissions", req.Permissions)
	utils.PrettyPrint(req)

	// Validate all permissions
	for _, permission := range req.Permissions {
		if !permissions.IsValidPermission(permission) {
			return apiErrors.ErrBadRequest().SetDetail("invalid permission: " + permission)
		}
	}

	// Check if user exists
	_, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("user not found")
		}
		return apiErrors.ErrInternalServerError().SetDetail("failed to verify user")
	}

	// Get current permissions to determine what to add/remove
	currentPerms, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), userID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to get current permissions")
	}

	slog.Info("Current permissions retrieved", "user_id", userID, "current_permissions_count", len(currentPerms))
	utils.PrettyPrint(currentPerms)

	// Safety check: if we can't find any current permissions but the user exists,
	// something might be wrong with the query. Don't proceed with bulk update.
	if len(currentPerms) == 0 && len(req.Permissions) == 0 {
		slog.Warn("No current permissions found and no new permissions requested - this might be a bug", "user_id", userID)
		return apiErrors.ErrBadRequest().SetDetail("No permissions to update")
	}

	// Additional safety: if we have no current permissions but are trying to assign new ones,
	// log this as it might indicate the GetUserPermissions query is not working
	if len(currentPerms) == 0 && len(req.Permissions) > 0 {
		slog.Warn("No current permissions found but new permissions requested - GetUserPermissions might be broken", 
			"user_id", userID, 
			"requested_permissions", req.Permissions)
	}

	// Create map of current permissions for easier lookup
	currentPermMap := make(map[string]bool)
	for _, perm := range currentPerms {
		currentPermMap[perm.PermissionID] = true
	}

	// Create map of new permissions
	newPermMap := make(map[string]bool)
	for _, perm := range req.Permissions {
		newPermMap[perm] = true
	}

	// Remove permissions that are no longer needed
	for _, currentPerm := range currentPerms {
		if !newPermMap[currentPerm.PermissionID] {
			slog.Info("Revoking permission", "user_id", userID, "permission_id", currentPerm.PermissionID)
			err := rg.gctx.Crate().Sqlite.Query().RevokeUserPermission(ctx.Context(), repository.RevokeUserPermissionParams{
				UserID:       userID,
				PermissionID: currentPerm.PermissionID,
			})
			if err != nil {
				return apiErrors.ErrInternalServerError().SetDetail("failed to revoke permission: " + currentPerm.PermissionID)
			}
		}
	}

	// Add new permissions
	for _, newPerm := range req.Permissions {
		if !currentPermMap[newPerm] {
			slog.Info("Assigning permission", "user_id", userID, "permission_id", newPerm)
			err := rg.gctx.Crate().Sqlite.Query().AssignUserPermission(ctx.Context(), repository.AssignUserPermissionParams{
				UserID:       userID,
				PermissionID: newPerm,
			})
			if err != nil {
				return apiErrors.ErrInternalServerError().SetDetail("failed to assign permission: " + newPerm)
			}
		}
	}

	return ctx.JSON(map[string]interface{}{
		"success":     true,
		"message":     "Permissions updated successfully",
		"user_id":     userID,
		"permissions": req.Permissions,
	})
}
