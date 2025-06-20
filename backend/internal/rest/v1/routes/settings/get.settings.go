package settings

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type SettingsResponse struct {
	RequestSystem    string `json:"request_system"`
	RequestSystemURL string `json:"request_system_url,omitempty"`
}

func (rg *RouteGroup) GetSettings(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	requestSystem, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingRequestSystem.String())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch request system setting")
	}
	requestSystemURL, err := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingRequestSystemURL.String())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch request system URL setting")
	}

	resp := SettingsResponse{
		RequestSystem: requestSystem,
	}
	if requestSystem == string(structures.RequestSystemExternal) {
		resp.RequestSystemURL = requestSystemURL
	}

	return ctx.JSON(resp)
}
