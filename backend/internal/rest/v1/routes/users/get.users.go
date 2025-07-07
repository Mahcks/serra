package users

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

// GetAllUsers returns all users with their basic info and permissions
func (rg *RouteGroup) GetAllUsers(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Get all users
	users, err := rg.gctx.Crate().Sqlite.Query().GetAllUsers(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch users")
	}

	// Build response with user info and permissions
	var userList []structures.UserWithPermissions
	for _, dbUser := range users {
		// Get user permissions
		userPermissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), dbUser.ID)
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user permissions")
		}

		// Convert to permission info - initialize as empty slice to ensure JSON returns []
		permissionInfos := make([]structures.PermissionInfo, 0)
		for _, userPerm := range userPermissions {
			permInfo := permissions.GetPermissionInfo(userPerm.PermissionID)
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

		userList = append(userList, structures.UserWithPermissions{
			ID:          dbUser.ID,
			Username:    dbUser.Username,
			Email:       email,
			AvatarUrl:   avatarUrl,
			UserType:    dbUser.UserType,
			CreatedAt:   createdAt,
			Permissions: permissionInfos,
		})
	}

	result := structures.GetAllUsersResponse{
		Total: int64(len(userList)),
		Users: userList,
	}

	return ctx.JSON(result)
}
