package settings

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// DefaultPermissionsResponse represents the API response for default permissions
type DefaultPermissionsResponse struct {
	Permissions map[string]bool `json:"permissions"`
}

// UpdateDefaultPermissionsRequest represents the API request for updating default permissions
type UpdateDefaultPermissionsRequest struct {
	Permissions map[string]bool `json:"permissions"`
}

// GetDefaultPermissions returns all available permissions and their default status
func (rg *RouteGroup) GetDefaultPermissions(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has owner permission
	userPerms, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), user.ID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user permissions")
	}

	hasOwnerPerm := false
	for _, perm := range userPerms {
		if perm.PermissionID == permissions.Owner {
			hasOwnerPerm = true
			break
		}
	}

	if !hasOwnerPerm {
		return apiErrors.ErrForbidden().SetDetail("owner permission required")
	}

	// Get default permission settings
	service := services.NewDynamicDefaultPermissionsService(rg.gctx.Crate().Sqlite.Query())
	settings, err := service.GetAllDefaultPermissionSettings(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch default permissions")
	}

	return ctx.JSON(DefaultPermissionsResponse{
		Permissions: settings,
	})
}

// UpdateDefaultPermissions updates which permissions are enabled by default for new users
func (rg *RouteGroup) UpdateDefaultPermissions(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has owner permission
	userPerms, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), user.ID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user permissions")
	}

	hasOwnerPerm := false
	for _, perm := range userPerms {
		if perm.PermissionID == permissions.Owner {
			hasOwnerPerm = true
			break
		}
	}

	if !hasOwnerPerm {
		return apiErrors.ErrForbidden().SetDetail("owner permission required")
	}

	var req UpdateDefaultPermissionsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	service := services.NewDynamicDefaultPermissionsService(rg.gctx.Crate().Sqlite.Query())

	// Update each permission setting
	for permissionID, enabled := range req.Permissions {
		// Validate that this is a real permission
		if !permissions.IsValidPermission(permissionID) {
			return apiErrors.ErrBadRequest().SetDetail("invalid permission: " + permissionID)
		}

		err := service.UpdateDefaultPermission(ctx.Context(), permissionID, enabled)
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update permission: " + permissionID)
		}
	}

	return ctx.JSON(map[string]string{"message": "Default permissions updated successfully"})
}