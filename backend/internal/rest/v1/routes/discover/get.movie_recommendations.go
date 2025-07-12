package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetMovieRecommendations(ctx *respond.Ctx) error {
	movieID := ctx.Params("movie_id")
	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.GetMovieRecommendations(movieID, page)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	return ctx.JSON(tmdbResp)
}
