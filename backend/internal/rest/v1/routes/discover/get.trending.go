package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetTrending(ctx *respond.Ctx) error {
	page := ctx.Query("page", "1")

	tmdbResp, err := rg.integrations.TMDB.GetTrendingMedia(page)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	// Get user from context for request/library status checking
	user := ctx.ParseClaims()
	if user == nil {
		// If no user context, return basic response without status
		return ctx.JSON(tmdbResp)
	}

	// For trending, we need to handle mixed media types (movies and TV)
	// We'll use "mixed" as the media type and handle it in the enrichment function
	enrichedResp, err := rg.enrichWithMediaStatus(ctx.Context(), &tmdbResp, user.ID, "mixed")
	if err != nil {
		// If enrichment fails, return basic response
		return ctx.JSON(tmdbResp)
	}

	return ctx.JSON(enrichedResp)
}
