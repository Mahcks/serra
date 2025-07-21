package analytics

import (
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetFailureAnalysis(c *respond.Ctx) error {
	user := c.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	// Get days parameter with default of 30
	daysStr := c.Query("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}

	// Calculate the date threshold
	since := time.Now().AddDate(0, 0, -days)

	// Get failure analysis from database
	failures, err := rg.gctx.Crate().Sqlite.Query().GetFailureAnalysis(c.Context(), since)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get failure analysis")
	}

	// If no data found, return empty array instead of null
	if failures == nil {
		failures = []repository.GetFailureAnalysisRow{}
	}

	return c.JSON(failures)
}