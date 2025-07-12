package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetTVWatchProviders(ctx *respond.Ctx) error {
	seriesID := ctx.Params("series_id")
	if seriesID == "" {
		return apiErrors.ErrBadRequest().SetDetail("series_id parameter is required")
	}

	result, err := rg.integrations.TMDB.GetTVWatchProviders(seriesID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch TV watch providers: %s", err)
	}

	return ctx.JSON(result)
}
