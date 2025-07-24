package settings

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

type AuthSettingsResponse struct {
	EnableMediaServerAuth    bool `json:"enable_media_server_auth"`
	EnableLocalAuth          bool `json:"enable_local_auth"`
	EnableNewMediaServerAuth bool `json:"enable_new_media_server_auth"`
}

func (rg *RouteGroup) GetAuthSettings(ctx *respond.Ctx) error {
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

	// Fetch authentication settings with defaults
	enableMediaServerAuth, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingEnableMediaServerAuth.String())
	enableLocalAuth, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingEnableLocalAuth.String())
	enableNewMediaServerAuth, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingEnableNewMediaServerAuth.String())

	// Set defaults if not configured
	if enableMediaServerAuth == "" {
		enableMediaServerAuth = "true" // Default: enabled
	}
	if enableLocalAuth == "" {
		enableLocalAuth = "false" // Default: disabled
	}
	if enableNewMediaServerAuth == "" {
		enableNewMediaServerAuth = "false" // Default: disabled
	}

	resp := AuthSettingsResponse{
		EnableMediaServerAuth:    enableMediaServerAuth == "true",
		EnableLocalAuth:          enableLocalAuth == "true",
		EnableNewMediaServerAuth: enableNewMediaServerAuth == "true",
	}

	return ctx.JSON(resp)
}