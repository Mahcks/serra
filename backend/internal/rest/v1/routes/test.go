package routes

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) TestRoute(ctx *respond.Ctx) error {
	// Get query parameters
	title := ctx.Query("title")
	yearStr := ctx.Query("year")
	mediaType := ctx.Query("type", "movie") // default to movie

	// Convert year from string to int
	var year int
	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			return apiErrors.ErrBadRequest().SetDetail("invalid year parameter")
		}
	}

	res, err := rg.integrations.RottenTomatoes.GetRottenTomatoesScore(ctx.Context(), title, year, mediaType)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get Rotten Tomatoes score")
	}

	return ctx.JSON(res)
}
