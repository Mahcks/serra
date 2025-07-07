package permissions

import (
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// RevokeUserPermission removes a permission from a user
func (rg *RouteGroup) RevokeUserPermission(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	permission := ctx.Params("permission")

	if userID == "" || permission == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID and permission are required")
	}

	// Validate permission exists
	if !permissions.IsValidPermission(permission) {
		return apiErrors.ErrBadRequest().SetDetail("invalid permission")
	}

	// Revoke permission
	err := rg.gctx.Crate().Sqlite.Query().RevokeUserPermission(ctx.Context(), repository.RevokeUserPermissionParams{
		UserID:       userID,
		PermissionID: permission,
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to revoke permission")
	}

	return ctx.JSON(map[string]interface{}{
		"success":    true,
		"message":    "Permission revoked successfully",
		"user_id":    userID,
		"permission": permission,
	})
}
