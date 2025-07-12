package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (r *RouteGroup) GetWatchProviders(ctx *respond.Ctx) error {
	// Get media type from query parameter (movie or tv)
	mediaType := ctx.Query("type", "movie")
	
	// Validate media type
	if mediaType != "movie" && mediaType != "tv" {
		return apiErrors.ErrBadRequest().SetDetail("Invalid media type. Must be 'movie' or 'tv'")
	}
	
	// Get watch providers for the specified media type
	providers, err := r.integrations.TMDB.GetWatchProviders(mediaType)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to fetch watch providers: " + err.Error())
	}
	
	return ctx.JSON(providers)
}

func (r *RouteGroup) GetWatchProviderRegions(ctx *respond.Ctx) error {
	// Get available regions for watch providers
	regions, err := r.integrations.TMDB.GetWatchProviderRegions()
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to fetch watch provider regions: " + err.Error())
	}
	
	return ctx.JSON(regions)
}