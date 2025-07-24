package settings

import (
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

type AuthSetting struct {
	Permission string `json:"permission"`
	Value      bool   `json:"value"`
}

type UpdateAuthSettingsRequest struct {
	Settings []AuthSetting `json:"settings"`
}

func (rg *RouteGroup) UpdateAuthSettings(ctx *respond.Ctx) error {
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

	var req UpdateAuthSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Validate that at least one auth method will remain enabled
	authStates := map[string]bool{
		"enable_media_server_auth":     true, // Current defaults
		"enable_local_auth":            false,
		"enable_new_media_server_auth": false,
	}

	// Apply the updates to check validation
	for _, setting := range req.Settings {
		switch setting.Permission {
		case "enable_media_server_auth", "enable_local_auth", "enable_new_media_server_auth":
			authStates[setting.Permission] = setting.Value
		default:
			return apiErrors.ErrBadRequest().SetDetail("invalid auth setting: " + setting.Permission)
		}
	}

	// Check that at least one auth method is enabled
	hasEnabledAuth := authStates["enable_media_server_auth"] || authStates["enable_local_auth"]
	if !hasEnabledAuth {
		return apiErrors.ErrBadRequest().SetDetail("at least one authentication method must be enabled")
	}

	// Update each setting in the database
	for _, setting := range req.Settings {
		var settingKey structures.Setting
		switch setting.Permission {
		case "enable_media_server_auth":
			settingKey = structures.SettingEnableMediaServerAuth
		case "enable_local_auth":
			settingKey = structures.SettingEnableLocalAuth
		case "enable_new_media_server_auth":
			settingKey = structures.SettingEnableNewMediaServerAuth
		}

		value := "false"
		if setting.Value {
			value = "true"
		}

		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   settingKey.String(),
			Value: value,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update " + setting.Permission)
		}
	}

	return ctx.JSON(map[string]string{"message": "Authentication settings updated successfully"})
}