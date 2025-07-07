package users

import (
	"database/sql"
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

// GetUser returns a specific user with their permissions
func (rg *RouteGroup) GetUser(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	// Get user details
	dbUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("user not found")
		}
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user")
	}

	utils.PrettyPrint(dbUser)

	// Get user permissions
	userPermissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), userID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user permissions")
	}

	utils.PrettyPrint(userPermissions)

	// Convert to permission info - initialize as empty slice to ensure JSON returns []
	permissionInfos := make([]structures.PermissionInfo, 0)
	for _, userPerm := range userPermissions {
		permInfo := permissions.GetPermissionInfo(userPerm.PermissionID)
		slog.Debug("Fetched permission info",
			"user_id", userID,
			"permission_id", userPerm.PermissionID,
			"permission_name", permInfo.Name,
			"permission_description", permInfo.Description)
		// Convert from permissions.PermissionInfo to structures.PermissionInfo
		permissionInfos = append(permissionInfos, structures.PermissionInfo{
			ID:          permInfo.ID,
			Name:        permInfo.Name,
			Description: permInfo.Description,
			Category:    permInfo.Category,
			Dangerous:   permInfo.Dangerous,
		})
	}

	email := ""
	if dbUser.Email.Valid {
		email = dbUser.Email.String
	}

	avatarUrl := ""
	if dbUser.AvatarUrl.Valid {
		avatarUrl = dbUser.AvatarUrl.String
	}

	createdAt := ""
	if dbUser.CreatedAt.Valid {
		createdAt = dbUser.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	userWithPermissions := structures.UserWithPermissions{
		ID:          dbUser.ID,
		Username:    dbUser.Username,
		Email:       email,
		AvatarUrl:   avatarUrl,
		UserType:    dbUser.UserType,
		CreatedAt:   createdAt,
		Permissions: permissionInfos,
	}

	return ctx.JSON(userWithPermissions)
}
