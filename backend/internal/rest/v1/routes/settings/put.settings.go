package settings

import (
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type UpdateSettingsRequest struct {
	RequestSystem       *string `json:"request_system,omitempty"`
	RequestSystemURL    *string `json:"request_system_url,omitempty"`
	JellystatEnabled    *bool   `json:"jellystat_enabled,omitempty"`
	JellystatHost       *string `json:"jellystat_host,omitempty"`
	JellystatPort       *string `json:"jellystat_port,omitempty"`
	JellystatUseSSL     *bool   `json:"jellystat_use_ssl,omitempty"`
	JellystatURL        *string `json:"jellystat_url,omitempty"`
	JellystatAPIKey     *string `json:"jellystat_api_key,omitempty"`
}

func (rg *RouteGroup) UpdateSettings(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	var req UpdateSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Update request system settings
	if req.RequestSystem != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingRequestSystem.String(),
			Value: *req.RequestSystem,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update request system setting")
		}
	}

	if req.RequestSystemURL != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingRequestSystemURL.String(),
			Value: *req.RequestSystemURL,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update request system URL setting")
		}
	}

	// Update Jellystat settings
	if req.JellystatEnabled != nil {
		value := "false"
		if *req.JellystatEnabled {
			value = "true"
		}
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatEnabled.String(),
			Value: value,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat enabled setting")
		}
	}

	if req.JellystatHost != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatHost.String(),
			Value: *req.JellystatHost,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat host setting")
		}
	}

	if req.JellystatPort != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatPort.String(),
			Value: *req.JellystatPort,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat port setting")
		}
	}

	if req.JellystatUseSSL != nil {
		value := "false"
		if *req.JellystatUseSSL {
			value = "true"
		}
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatUseSSL.String(),
			Value: value,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat SSL setting")
		}
	}

	if req.JellystatURL != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatURL.String(),
			Value: *req.JellystatURL,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat URL setting")
		}
	}

	if req.JellystatAPIKey != nil {
		err := rg.gctx.Crate().Sqlite.Query().UpsertSetting(ctx.Context(), repository.UpsertSettingParams{
			Key:   structures.SettingJellystatAPIKey.String(),
			Value: *req.JellystatAPIKey,
		})
		if err != nil {
			return apiErrors.ErrInternalServerError().SetDetail("failed to update Jellystat API key setting")
		}
	}

	return ctx.JSON(map[string]string{"message": "Settings updated successfully"})
}