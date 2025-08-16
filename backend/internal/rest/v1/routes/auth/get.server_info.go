package auth

import (
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/structures"
)

type ServerInfoResponse struct {
	MediaServerType string `json:"media_server_type"`
	MediaServerName string `json:"media_server_name"`
	LocalAuthEnabled bool   `json:"local_auth_enabled"`
	MediaServerAuthEnabled bool `json:"media_server_auth_enabled"`
}

// GetServerInfo returns authentication configuration information
func (rg *RouteGroup) GetServerInfo(ctx *respond.Ctx) error {
	// Get media server type
	mediaServerType := string(rg.Config().MediaServer.Type)
	
	// Determine display name based on type
	var mediaServerName string
	switch rg.Config().MediaServer.Type {
	case structures.ProviderJellyfin:
		mediaServerName = "Jellyfin"
	case structures.ProviderEmby:
		mediaServerName = "Emby"
	default:
		mediaServerName = "Media Server"
	}

	// Check if local auth is enabled
	localAuthEnabled, err := rg.checkAuthMethodEnabled(ctx, structures.SettingEnableLocalAuth.String())
	if err != nil {
		slog.Error("Failed to check local auth enabled", "error", err)
		localAuthEnabled = true // Default to enabled for safety
	}

	// Check if media server auth is enabled
	mediaServerAuthEnabled, err := rg.checkAuthMethodEnabled(ctx, structures.SettingEnableMediaServerAuth.String())
	if err != nil {
		slog.Error("Failed to check media server auth enabled", "error", err)
		mediaServerAuthEnabled = true // Default to enabled for safety
	}

	response := ServerInfoResponse{
		MediaServerType:        mediaServerType,
		MediaServerName:        mediaServerName,
		LocalAuthEnabled:       localAuthEnabled,
		MediaServerAuthEnabled: mediaServerAuthEnabled,
	}

	return ctx.JSON(response)
}