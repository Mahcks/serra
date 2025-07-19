package discover

import (
	"context"
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/season_availability"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetSeasonAvailability(ctx *respond.Ctx) error {
	tmdbIDStr := ctx.Params("id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("TMDB ID must be a valid integer")
	}

	// Get season availability service
	seasonService := season_availability.NewSeasonAvailabilityService(rg.gctx.Crate().Sqlite.Query(), rg.integrations.Emby)
	
	availability, err := seasonService.GetSeasonAvailability(context.Background(), tmdbID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get season availability: " + err.Error())
	}

	return ctx.JSON(availability)
}

func (rg *RouteGroup) SyncSeasonAvailability(ctx *respond.Ctx) error {
	tmdbIDStr := ctx.Params("id")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("TMDB ID must be a valid integer")
	}

	// Get season availability service
	seasonService := season_availability.NewSeasonAvailabilityService(rg.gctx.Crate().Sqlite.Query(), rg.integrations.Emby)
	
	err = seasonService.SyncShowAvailability(context.Background(), tmdbID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to sync season availability: " + err.Error())
	}

	return ctx.JSON(map[string]interface{}{
		"message": "Season availability synced successfully",
		"tmdb_id": tmdbID,
	})
}