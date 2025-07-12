package discover

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetDiscoverMovie(ctx *respond.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page"))

	params := structures.DiscoverMovieParams{
		Page:                       page,
		WithGenres:                 ctx.Query("with_genres"),
		ReleaseDateGTE:             ctx.Query("release_date.gte"),
		ReleaseDateLTE:             ctx.Query("release_date.lte"),
		WithCompanies:              ctx.Query("with_companies"),
		WithKeywords:               ctx.Query("with_keywords"),
		WithOriginalLanguage:       ctx.Query("with_original_language"),
		WithRuntimeGTE:             atoi(ctx.Query("with_runtime.gte")),
		WithRuntimeLTE:             atoi(ctx.Query("with_runtime.lte")),
		VoteAverageGTE:             atof(ctx.Query("vote_average.gte")),
		VoteAverageLTE:             atof(ctx.Query("vote_average.lte")),
		VoteCountGTE:               atoi(ctx.Query("vote_count.gte")),
		VoteCountLTE:               atoi(ctx.Query("vote_count.lte")),
		WithWatchProviders:         ctx.Query("with_watch_providers"),
		WithWatchMonetizationTypes: ctx.Query("with_watch_monetization_types"),
		WatchRegion:                ctx.Query("watch_region"),
		SortBy:                     ctx.Query("sort_by", "popularity.desc"),
	}

	tmdbResp, err := rg.integrations.TMDB.DiscoverMovie(params)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(
			"failed to discover movies: %s", err,
		)
	}

	return ctx.JSON(tmdbResp)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
