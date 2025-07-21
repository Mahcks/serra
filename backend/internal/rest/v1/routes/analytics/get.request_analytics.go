package analytics

import (
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

// GetRequestAnalytics returns comprehensive request/fulfillment analytics
func (rg *RouteGroup) GetRequestAnalytics(ctx *respond.Ctx) error {
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

	limitParam := ctx.Query("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Calculate date range
	since := time.Now().AddDate(0, 0, -days)

	// Get request success rates
	successRates, err := rg.gctx.Crate().Sqlite.Query().GetRequestSuccessRates(ctx.Context(), since)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get success rates")
	}

	// Get request fulfillment by user
	userFulfillment, err := rg.gctx.Crate().Sqlite.Query().GetRequestFulfillmentByUser(ctx.Context(), repository.GetRequestFulfillmentByUserParams{
		CreatedAt: since,
		Limit:     int64(limit),
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get user fulfillment data")
	}

	// Get popular requested content
	popularContent, err := rg.gctx.Crate().Sqlite.Query().GetPopularRequestedContent(ctx.Context(), repository.GetPopularRequestedContentParams{
		CreatedAt: since,
		Limit:     int64(limit),
	})
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get popular content")
	}

	// Get processing performance
	performance, err := rg.gctx.Crate().Sqlite.Query().GetRequestProcessingPerformance(ctx.Context(), since)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get processing performance")
	}

	// Get failure analysis
	failures, err := rg.gctx.Crate().Sqlite.Query().GetFailureAnalysis(ctx.Context(), since)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get failure analysis")
	}

	return ctx.JSON(map[string]interface{}{
		"success_rates":     successRates,
		"user_fulfillment":  userFulfillment,
		"popular_content":   popularContent,
		"performance":       performance,
		"failure_analysis":  failures,
		"period_days":       days,
	})
}