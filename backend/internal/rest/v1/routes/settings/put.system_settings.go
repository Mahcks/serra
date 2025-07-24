package settings

import (
	"strconv"
	
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

type UpdateSystemSettingsRequest struct {
	Settings map[string]interface{} `json:"settings"`
}


func (rg *RouteGroup) UpdateSystemSettings(ctx *respond.Ctx) error {
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

	var req UpdateSystemSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Validate authentication settings if they're being updated
	authSettings := map[string]bool{
		"enable_media_server_auth": true, // Current defaults
		"enable_local_auth":        false,
	}

	// Check if any auth settings are being updated and validate them
	hasAuthUpdates := false
	for settingName, value := range req.Settings {
		if settingName == "enable_media_server_auth" || settingName == "enable_local_auth" {
			hasAuthUpdates = true
			if boolVal, ok := value.(bool); ok {
				authSettings[settingName] = boolVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("authentication settings must be boolean values")
			}
		}
	}

	// If auth settings are being updated, ensure at least one method remains enabled
	if hasAuthUpdates {
		hasEnabledAuth := authSettings["enable_media_server_auth"] || authSettings["enable_local_auth"]
		if !hasEnabledAuth {
			return apiErrors.ErrBadRequest().SetDetail("at least one authentication method must be enabled")
		}
	}

	// Update each setting in the database
	for settingName, value := range req.Settings {
		var settingKey structures.Setting
		var stringValue string

		// Map setting names to their constants and convert values to strings
		switch settingName {
		case "request_system":
			settingKey = structures.SettingRequestSystem
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("request_system must be a string")
			}
		case "request_system_url":
			settingKey = structures.SettingRequestSystemURL
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("request_system_url must be a string")
			}
		case "enable_media_server_auth":
			settingKey = structures.SettingEnableMediaServerAuth
			if boolVal, ok := value.(bool); ok {
				stringValue = "false"
				if boolVal {
					stringValue = "true"
				}
			} else {
				return apiErrors.ErrBadRequest().SetDetail("enable_media_server_auth must be a boolean")
			}
		case "enable_local_auth":
			settingKey = structures.SettingEnableLocalAuth
			if boolVal, ok := value.(bool); ok {
				stringValue = "false"
				if boolVal {
					stringValue = "true"
				}
			} else {
				return apiErrors.ErrBadRequest().SetDetail("enable_local_auth must be a boolean")
			}
		case "jellystat_enabled":
			settingKey = structures.SettingJellystatEnabled
			if boolVal, ok := value.(bool); ok {
				stringValue = "false"
				if boolVal {
					stringValue = "true"
				}
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_enabled must be a boolean")
			}
		case "jellystat_host":
			settingKey = structures.SettingJellystatHost
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_host must be a string")
			}
		case "jellystat_port":
			settingKey = structures.SettingJellystatPort
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_port must be a string")
			}
		case "jellystat_use_ssl":
			settingKey = structures.SettingJellystatUseSSL
			if boolVal, ok := value.(bool); ok {
				stringValue = "false"
				if boolVal {
					stringValue = "true"
				}
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_use_ssl must be a boolean")
			}
		case "jellystat_url":
			settingKey = structures.SettingJellystatURL
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_url must be a string")
			}
		case "jellystat_api_key":
			settingKey = structures.SettingJellystatAPIKey
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("jellystat_api_key must be a string")
			}
		case "tmdb_api_key":
			settingKey = structures.SettingTMDBAPIKey
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("tmdb_api_key must be a string")
			}
		case "download_visibility":
			settingKey = structures.SettingDownloadVisibility
			if strVal, ok := value.(string); ok {
				stringValue = strVal
			} else {
				return apiErrors.ErrBadRequest().SetDetail("download_visibility must be a string")
			}
		case "global_movie_request_limit":
			settingKey = structures.SettingGlobalMovieRequestLimit
			if intVal, ok := value.(float64); ok { // JSON numbers come as float64
				stringValue = strconv.Itoa(int(intVal))
			} else {
				return apiErrors.ErrBadRequest().SetDetail("global_movie_request_limit must be a number")
			}
		case "global_series_request_limit":
			settingKey = structures.SettingGlobalSeriesRequestLimit
			if intVal, ok := value.(float64); ok { // JSON numbers come as float64
				stringValue = strconv.Itoa(int(intVal))
			} else {
				return apiErrors.ErrBadRequest().SetDetail("global_series_request_limit must be a number")
			}
		default:
			return apiErrors.ErrBadRequest().SetDetail("unknown setting: " + settingName)
		}

		// Update the setting in the database
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   settingKey.String(),
			Value: stringValue,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update " + settingName)
		}
	}

	return ctx.JSON(map[string]string{"message": "System settings updated successfully"})
}