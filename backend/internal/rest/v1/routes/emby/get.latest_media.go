package emby

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetLatestMedia(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Fetch latest media items from Emby
	latestMedia, err := rg.integrations.Emby.GetLatestMedia(user)
	if err != nil {
		return err
	}

	return ctx.JSON(latestMedia)
}
