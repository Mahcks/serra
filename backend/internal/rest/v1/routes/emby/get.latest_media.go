package emby

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetLatestMedia(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Get media server type from settings to determine which integration to use
	mediaServerType := rg.gctx.Crate().Config.Get().MediaServer.Type

	// Fetch latest media items from the configured media server
	var latestMedia interface{}
	var err error

	switch mediaServerType {
	case structures.ProviderEmby:
		latestMedia, err = rg.integrations.Emby.GetLatestMedia(user)
	case structures.ProviderJellyfin:
		// For now, Jellyfin uses the same API structure as Emby, so we can reuse the Emby integration
		// In the future, you might want to create a separate Jellyfin integration
		latestMedia, err = rg.integrations.Emby.GetLatestMedia(user)
	default:
		return apiErrors.ErrBadRequest().SetDetail("unsupported media server type")
	}

	if err != nil {
		return err
	}

	return ctx.JSON(latestMedia)
}
