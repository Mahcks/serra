package analytics

import (
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetContentAvailability(c *respond.Ctx) error {
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

	// Get limit parameter with default of 10
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Calculate the date threshold
	since := time.Now().AddDate(0, 0, -days)

	// Get content availability vs requests from database
	params := repository.GetContentAvailabilityVsRequestsParams{
		CreatedAt: since,
		Limit:     int64(limit),
	}

	availability, err := rg.gctx.Crate().Sqlite.Query().GetContentAvailabilityVsRequests(c.Context(), params)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get content availability")
	}

	// If no data found, return empty array instead of null
	if availability == nil {
		availability = []repository.GetContentAvailabilityVsRequestsRow{}
	}

	return c.JSON(availability)
}