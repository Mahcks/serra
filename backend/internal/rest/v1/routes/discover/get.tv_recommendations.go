package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetTVRecommendations(ctx *respond.Ctx) error {
	seriesID := ctx.Params("series_id")
	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.GetTvRecommendations(seriesID, page)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	return ctx.JSON(tmdbResp)
}
