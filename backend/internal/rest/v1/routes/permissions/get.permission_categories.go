package permissions

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
)

// GetPermissionsByCategory returns permissions grouped by category
func (rg *RouteGroup) GetPermissionsByCategory(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	categories := permissions.GetPermissionsByCategory()
	
	// Build response with full permission info for each category
	response := make(map[string][]permissions.PermissionInfo)
	
	for category, permList := range categories {
		var categoryPerms []permissions.PermissionInfo
		for _, perm := range permList {
			categoryPerms = append(categoryPerms, permissions.GetPermissionInfo(perm))
		}
		response[category] = categoryPerms
	}
	
	return ctx.JSON(response)
}