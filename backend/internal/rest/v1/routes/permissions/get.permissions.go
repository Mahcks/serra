package permissions

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

// GetAllPermissions returns all available permissions for admin UI
func (rg *RouteGroup) GetAllPermissions(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Return all permission info for admin UI
	allPermissions := permissions.GetAllPermissionInfo()

	if len(allPermissions) == 0 {
		return apiErrors.ErrNotFound().SetDetail("No permissions found")
	}

	// Prepare response with all permissions and categorized permissions
	// This assumes GetPermissionsByCategory() returns a map[string][]PermissionInfo
	permissions := permissions.GetPermissionsByCategory()
	if len(permissions) == 0 {
		return apiErrors.ErrNotFound().SetDetail("No categorized permissions found")
	}

	// Prepare the response structure
	perms := make([]structures.PermissionInfo, 0, len(allPermissions))
	for _, perm := range allPermissions {
		perms = append(perms, structures.PermissionInfo{
			ID:          perm.ID,
			Name:        perm.Name,
			Description: perm.Description,
			Category:    perm.Category,
			Dangerous:   perm.Dangerous,
		})
	}

	response := structures.PermissionsListResponse{
		Permissions: perms,
		Categories:  permissions,
	}
	return ctx.JSON(response)
}
