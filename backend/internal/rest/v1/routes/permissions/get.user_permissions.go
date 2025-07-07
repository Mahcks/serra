package permissions

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// GetUserPermissions returns all permissions for a specific user
func (rg *RouteGroup) GetUserPermissions(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	// Get user permissions from database
	userPermissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), userID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user permissions")
	}

	// Convert to permission info with details
	var permissionInfos []permissions.PermissionInfo
	for _, userPerm := range userPermissions {
		permInfo := permissions.GetPermissionInfo(userPerm.PermissionID)
		permissionInfos = append(permissionInfos, permInfo)
	}

	return ctx.JSON(map[string]interface{}{
		"user_id":     userID,
		"permissions": permissionInfos,
	})
}
