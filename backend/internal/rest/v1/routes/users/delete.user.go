package users

import (
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

// DeleteUser deletes a user and all associated data
func (rg *RouteGroup) DeleteUser(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	// Prevent users from deleting themselves
	if userID == user.ID {
		return apiErrors.ErrBadRequest().SetDetail("cannot delete your own account")
	}

	// Check if user exists
	exists, err := rg.gctx.Crate().Sqlite.Query().UserExists(ctx.Context(), userID)
	if err != nil {
		slog.Error("Failed to check if user exists", "error", err, "user_id", userID)
		return apiErrors.ErrInternalServerError().SetDetail("failed to check user existence")
	}

	if exists == 0 {
		return apiErrors.ErrNotFound().SetDetail("user not found")
	}

	// Delete user permissions first (foreign key constraint)
	err = rg.gctx.Crate().Sqlite.Query().DeleteUserPermissions(ctx.Context(), userID)
	if err != nil {
		slog.Error("Failed to delete user permissions", "error", err, "user_id", userID)
		return apiErrors.ErrInternalServerError().SetDetail("failed to delete user permissions")
	}

	// Delete the user
	err = rg.gctx.Crate().Sqlite.Query().DeleteUser(ctx.Context(), userID)
	if err != nil {
		slog.Error("Failed to delete user", "error", err, "user_id", userID)
		return apiErrors.ErrInternalServerError().SetDetail("failed to delete user")
	}

	slog.Info("User deleted successfully", "user_id", userID, "deleted_by", user.ID)

	return ctx.JSON(map[string]interface{}{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}