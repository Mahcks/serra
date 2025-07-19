package discover

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetDiscoverTV(ctx *respond.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page"))

	params := structures.DiscoverTVParams{
		Page: page,

		// Date range filters
		AirDateGTE:       ctx.Query("air_date.gte"),
		AirDateLTE:       ctx.Query("air_date.lte"),
		FirstAirDateYear: atoi(ctx.Query("first_air_date_year")),
		FirstAirDateGTE:  ctx.Query("first_air_date.gte"),
		FirstAirDateLTE:  ctx.Query("first_air_date.lte"),

		// Content filters
		IncludeAdult:             ctx.Query("include_adult") == "true",
		IncludeNullFirstAirDates: ctx.Query("include_null_first_air_dates") == "true",
		Language:                 ctx.Query("language"),
		ScreenedTheatrically:     ctx.Query("screened_theatrically") == "true",
		Timezone:                 ctx.Query("timezone"),

		// Rating filters
		VoteAverageGTE: atof(ctx.Query("vote_average.gte")),
		VoteAverageLTE: atof(ctx.Query("vote_average.lte")),
		VoteCountGTE:   atof(ctx.Query("vote_count.gte")),
		VoteCountLTE:   atof(ctx.Query("vote_count.lte")),

		// Content categorization
		WithGenres:           ctx.Query("with_genres"),
		WithKeywords:         ctx.Query("with_keywords"),
		WithCompanies:        ctx.Query("with_companies"),
		WithNetworks:         atoi(ctx.Query("with_networks")),
		WithOriginCountry:    ctx.Query("with_origin_country"),
		WithOriginalLanguage: ctx.Query("with_original_language"),

		// Runtime filters
		WithRuntimeGTE: atoi(ctx.Query("with_runtime.gte")),
		WithRuntimeLTE: atoi(ctx.Query("with_runtime.lte")),

		// Status filters
		WithStatus: ctx.Query("with_status"),

		// Watch providers
		WatchRegion:                ctx.Query("watch_region"),
		WithWatchMonetizationTypes: ctx.Query("with_watch_monetization_types"),
		WithWatchProviders:         ctx.Query("with_watch_providers"),

		// Exclusion filters
		WithoutCompanies:      ctx.Query("without_companies"),
		WithoutGenres:         ctx.Query("without_genres"),
		WithoutKeywords:       ctx.Query("without_keywords"),
		WithoutWatchProviders: ctx.Query("without_watch_providers"),

		// Type filter
		WithType: ctx.Query("with_type"),

		// Sorting
		SortBy: ctx.Query("sort_by", "popularity.desc"),
	}

	tmdbResp, err := rg.integrations.TMDB.DiscoverTV(params)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to discover TV shows: %s", err,
		)
	}

	// Get user from context for request/library status checking
	user := ctx.ParseClaims()
	if user == nil {
		// If no user context, return basic response without status
		return ctx.JSON(tmdbResp)
	}

	// Enrich response with request and library status
	enrichedResp, err := rg.enrichWithMediaStatus(ctx.Context(), &tmdbResp, user.ID, "tv")
	if err != nil {
		// If enrichment fails, return basic response
		return ctx.JSON(tmdbResp)
	}

	return ctx.JSON(enrichedResp)
}