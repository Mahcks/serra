package permissions

import (
	"database/sql"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

// AssignUserPermission assigns a permission to a user
func (rg *RouteGroup) AssignUserPermission(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	var req structures.AssignPermissionRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Validate permission exists
	if !permissions.IsValidPermission(req.Permission) {
		return apiErrors.ErrBadRequest().SetDetail("invalid permission")
	}

	// Check if user exists
	_, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("user not found")
		}
		return apiErrors.ErrInternalServerError().SetDetail("failed to verify user")
	}

	// Assign permission
	err = rg.gctx.Crate().Sqlite.Query().AssignUserPermission(ctx.Context(), repository.AssignUserPermissionParams{
		UserID:       userID,
		PermissionID: req.Permission,
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to assign permission")
	}

	return ctx.JSON(map[string]interface{}{
		"success":    true,
		"message":    "Permission assigned successfully",
		"user_id":    userID,
		"permission": req.Permission,
	})
}
