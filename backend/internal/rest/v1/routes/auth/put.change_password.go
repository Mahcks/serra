package auth

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// ChangeLocalUserPassword changes the password for a local user
func (rg *RouteGroup) ChangeLocalUserPassword(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	targetUserID := ctx.Params("id")
	if targetUserID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	// Check if user has owner OR admin.users permission to change passwords
	hasOwnerPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
		UserID:       user.ID,
		PermissionID: "owner",
	})
	if err != nil {
		hasOwnerPermission = false
	}

	hasAdminUsersPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
		UserID:       user.ID,
		PermissionID: "admin.users",
	})
	if err != nil {
		hasAdminUsersPermission = false
	}

	// Allow users to change their own password OR require admin permissions
	isSelfPasswordChange := user.ID == targetUserID
	if !isSelfPasswordChange && !hasOwnerPermission && !hasAdminUsersPermission {
		return apiErrors.ErrForbidden().SetDetail("Missing required permission: owner or admin.users")
	}

	var req structures.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Get target user to verify they exist and are local
	targetUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), targetUserID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("User not found")
	}

	// Only allow password changes for local users
	if targetUser.UserType != "local" {
		return apiErrors.ErrBadRequest().SetDetail("Password can only be changed for local users")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to hash password")
	}

	// Update password
	err = rg.gctx.Crate().Sqlite.Query().UpdateUserPassword(ctx.Context(), repository.UpdateUserPasswordParams{
		ID:           targetUserID,
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to update password")
	}

	return ctx.JSON(map[string]interface{}{
		"message": "Password updated successfully",
	})
}
