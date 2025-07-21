package analytics

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

// GetDriveAlerts returns all active drive alerts for admin monitoring
func (rg *RouteGroup) GetDriveAlerts(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	alerts, err := rg.driveMonitor.GetActiveDriveAlerts(ctx.Context())
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to get drive alerts")
	}

	return ctx.JSON(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}