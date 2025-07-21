package analytics

import (
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

// GetRequestTrends returns request trends over time
func (rg *RouteGroup) GetRequestTrends(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	// Parse query parameters
	daysParam := ctx.Query("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days < 1 {
		days = 30
	}

	limitParam := ctx.Query("limit", "30")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = 30
	}

	// Calculate date range
	since := time.Now().AddDate(0, 0, -days)

	// Get request trends
	trends, err := rg.gctx.Crate().Sqlite.Query().GetRequestTrends(ctx.Context(), repository.GetRequestTrendsParams{
		CreatedAt: since,
		Limit:     int64(limit),
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get request trends")
	}

	// Get request volume by hour
	hourlyVolume, err := rg.gctx.Crate().Sqlite.Query().GetRequestVolumeByHour(ctx.Context(), since)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get hourly volume")
	}

	return ctx.JSON(map[string]interface{}{
		"daily_trends":   trends,
		"hourly_volume":  hourlyVolume,
		"period_days":    days,
	})
}