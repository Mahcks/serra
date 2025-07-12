package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) SearchMovie(ctx *respond.Ctx) error {
	query := ctx.Query("query")
	if query == "" {
		return apiErrors.ErrBadRequest().SetDetail("query parameter is required")
	}

	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.SearchMovie(query, page)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to search movie: %s", err,
		)
	}

	return ctx.JSON(tmdbResp)
}
