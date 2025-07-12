package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetMovieWatchProviders(ctx *respond.Ctx) error {
	movieID := ctx.Params("movie_id")

	result, err := rg.integrations.TMDB.GetMovieWatchProviders(movieID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to fetch movie watch providers: %s", err,
		)
	}

	return ctx.JSON(result)
}
