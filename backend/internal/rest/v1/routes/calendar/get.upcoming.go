package calendar

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

func (rg *RouteGroup) GetUpcomingMedia(ctx *respond.Ctx) error {
	var result []structures.CalendarItem

	// Fetch wanted items from Radarr
	radarrItems, err := rg.gctx.Crate().Radarr.GetCalendarItems(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch Radarr calendar items")
	}
	result = append(result, radarrItems...)

	// Fetch wanted items from Sonarr
	sonarrItems, err := rg.gctx.Crate().Sonarr.GetUpcomingItems(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to fetch Sonarr calendar items")
	}
	result = append(result, sonarrItems...)

	return ctx.JSON(result)
}
