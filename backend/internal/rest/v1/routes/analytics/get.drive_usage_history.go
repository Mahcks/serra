package analytics

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

// GetDriveUsageHistory returns historical usage data for a specific drive
func (rg *RouteGroup) GetDriveUsageHistory(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	driveID := ctx.Params("driveId")
	if driveID == "" {
		return apiErrors.ErrBadRequest().SetDetail("Drive ID is required")
	}

	// Parse optional days parameter (default to 30 days)
	days := 30
	if daysStr := ctx.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	history, err := rg.driveMonitor.GetDriveUsageHistory(ctx.Context(), driveID, days)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get drive usage history")
	}

	return ctx.JSON(map[string]interface{}{
		"drive_id": driveID,
		"days":     days,
		"history":  history,
		"count":    len(history),
	})
}