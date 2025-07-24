package settings

import (
	"strconv"
	
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/permissions"
	"github.com/mahcks/serra/pkg/structures"
)

type SystemSettingsResponse struct {
	// Request system settings
	RequestSystem    string `json:"request_system"`
	RequestSystemURL string `json:"request_system_url,omitempty"`
	
	// Authentication settings
	EnableMediaServerAuth bool `json:"enable_media_server_auth"`
	EnableLocalAuth       bool `json:"enable_local_auth"`
	
	// Jellystat settings
	JellystatEnabled bool   `json:"jellystat_enabled"`
	JellystatHost    string `json:"jellystat_host,omitempty"`
	JellystatPort    string `json:"jellystat_port,omitempty"`
	JellystatUseSSL  bool   `json:"jellystat_use_ssl"`
	JellystatURL     string `json:"jellystat_url,omitempty"`
	JellystatAPIKey  string `json:"jellystat_api_key,omitempty"`
	
	// TMDB settings
	TMDBAPIKey string `json:"tmdb_api_key,omitempty"`
	
	// Download visibility
	DownloadVisibility string `json:"download_visibility"`
	
	// Request limits
	GlobalMovieRequestLimit  int `json:"global_movie_request_limit"`
	GlobalSeriesRequestLimit int `json:"global_series_request_limit"`
	
}

func (rg *RouteGroup) GetSystemSettings(ctx *respond.Ctx) error {
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

	// Fetch all system settings
	requestSystem, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingRequestSystem.String())
	requestSystemURL, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingRequestSystemURL.String())
	
	// Authentication settings with defaults
	enableMediaServerAuth, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingEnableMediaServerAuth.String())
	enableLocalAuth, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingEnableLocalAuth.String())
	
	// Jellystat settings
	jellystatEnabled, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatEnabled.String())
	jellystatHost, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatHost.String())
	jellystatPort, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatPort.String())
	jellystatUseSSL, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatUseSSL.String())
	jellystatURL, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatURL.String())
	jellystatAPIKey, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingJellystatAPIKey.String())
	
	// TMDB settings
	tmdbAPIKey, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingTMDBAPIKey.String())
	
	// Download visibility
	downloadVisibility, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingDownloadVisibility.String())
	
	// Request limits
	globalMovieRequestLimit, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingGlobalMovieRequestLimit.String())
	globalSeriesRequestLimit, _ := rg.gctx.Crate().Sqlite.Query().GetSetting(ctx.Context(), structures.SettingGlobalSeriesRequestLimit.String())
	

	// Set defaults for settings that don't have values
	if requestSystem == "" {
		requestSystem = string(structures.RequestSystemBuiltIn)
	}
	if enableMediaServerAuth == "" {
		enableMediaServerAuth = "true" // Default: enabled
	}
	if enableLocalAuth == "" {
		enableLocalAuth = "false" // Default: disabled
	}
	if downloadVisibility == "" {
		downloadVisibility = string(structures.DownloadVisibilityAll)
	}
	
	// Convert request limits to integers with defaults
	movieRequestLimit := 0 // Default: unlimited
	if globalMovieRequestLimit != "" {
		if limit, err := strconv.Atoi(globalMovieRequestLimit); err == nil {
			movieRequestLimit = limit
		}
	}
	
	seriesRequestLimit := 0 // Default: unlimited
	if globalSeriesRequestLimit != "" {
		if limit, err := strconv.Atoi(globalSeriesRequestLimit); err == nil {
			seriesRequestLimit = limit
		}
	}
	

	resp := SystemSettingsResponse{
		RequestSystem:         requestSystem,
		RequestSystemURL:      requestSystemURL,
		EnableMediaServerAuth: enableMediaServerAuth == "true",
		EnableLocalAuth:       enableLocalAuth == "true",
		JellystatEnabled:         jellystatEnabled == "true",
		JellystatHost:            jellystatHost,
		JellystatPort:            jellystatPort,
		JellystatUseSSL:          jellystatUseSSL == "true",
		JellystatURL:             jellystatURL,
		JellystatAPIKey:          jellystatAPIKey,
		TMDBAPIKey:               tmdbAPIKey,
		DownloadVisibility:       downloadVisibility,
		GlobalMovieRequestLimit:  movieRequestLimit,
		GlobalSeriesRequestLimit: seriesRequestLimit,
	}

	return ctx.JSON(resp)
}