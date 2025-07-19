package discover

import (
	"fmt"
	"net/http"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetSeasonDetails(ctx *respond.Ctx) error {
	seriesID := ctx.Params("series_id")
	seasonNumber := ctx.Params("season_number")

	if seriesID == "" || seasonNumber == "" {
		return apiErrors.ErrBadRequest().SetDetail("Missing series_id or season_number")
	}

	// Get season details from TMDB
	seasonDetails, err := rg.integrations.TMDB.GetSeasonDetails(seriesID, seasonNumber)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail(fmt.Sprintf("failed to fetch season details: %s", err))
	}

	return ctx.Status(http.StatusOK).JSON(seasonDetails)
}