package settings

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type SettingsResponse struct {
	RequestSystem              string `json:"request_system"`
	RequestSystemURL           string `json:"request_system_url,omitempty"`
	JellystatEnabled           bool   `json:"jellystat_enabled"`
	JellystatHost              string `json:"jellystat_host,omitempty"`
	JellystatPort              string `json:"jellystat_port,omitempty"`
	JellystatUseSSL            bool   `json:"jellystat_use_ssl"`
	JellystatURL               string `json:"jellystat_url,omitempty"`
	JellystatAPIKey            string `json:"jellystat_api_key,omitempty"`
	EnableMediaServerAuth      bool   `json:"enable_media_server_auth"`
	EnableLocalAuth            bool   `json:"enable_local_auth"`
	EnableNewMediaServerAuth   bool   `json:"enable_new_media_server_auth"`
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

	// Fetch Jellystat settings
	jellystatEnabled, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatEnabled.String())
	jellystatHost, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatHost.String())
	jellystatPort, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatPort.String())
	jellystatUseSSL, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatUseSSL.String())
	jellystatURL, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatURL.String())
	jellystatAPIKey, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatAPIKey.String())

	// Fetch authentication settings (with defaults)
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

	resp := SettingsResponse{
		RequestSystem:            requestSystem,
		JellystatEnabled:         jellystatEnabled == "true",
		JellystatHost:            jellystatHost,
		JellystatPort:            jellystatPort,
		JellystatUseSSL:          jellystatUseSSL == "true",
		JellystatURL:             jellystatURL,
		JellystatAPIKey:          jellystatAPIKey,
		EnableMediaServerAuth:    enableMediaServerAuth == "true",
		EnableLocalAuth:          enableLocalAuth == "true",
		EnableNewMediaServerAuth: enableNewMediaServerAuth == "true",
	}
	if requestSystem == string(structures.RequestSystemExternal) {
		resp.RequestSystemURL = requestSystemURL
	}

	return ctx.JSON(resp)
}
