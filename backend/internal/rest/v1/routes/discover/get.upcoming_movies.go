package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetUpcomingMovies(ctx *respond.Ctx) error {
	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.GetMovieUpcoming(page)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to fetch upcoming movies: %s", err,
		)
	}

	// Get user from context for request/library status checking
	user := ctx.ParseClaims()
	if user == nil {
		// If no user context, return basic response without status
		return ctx.JSON(tmdbResp)
	}

	// Enrich response with request and library status
	enrichedResp, err := rg.enrichWithMediaStatus(ctx.Context(), &tmdbResp, user.ID, "movie")
	if err != nil {
		// If enrichment fails, return basic response
		return ctx.JSON(tmdbResp)
	}

	return ctx.JSON(enrichedResp)
}