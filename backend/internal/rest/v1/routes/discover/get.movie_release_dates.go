package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetMovieReleaseDates(ctx *respond.Ctx) error {
	movieID := ctx.Params("movie_id")

	tmdbResp, err := rg.integrations.TMDB.GetMovieReleaseDates(movieID)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	return ctx.JSON(tmdbResp)
}