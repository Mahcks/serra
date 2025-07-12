package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetPopularTV(ctx *respond.Ctx) error {
	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.GetTVPopular(page)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to fetch popular TV shows: %s", err,
		)
	}

	return ctx.JSON(tmdbResp)
}