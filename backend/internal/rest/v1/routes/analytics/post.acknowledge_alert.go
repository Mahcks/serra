package analytics

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/pkg/api_errors"
)

// AcknowledgeAlert marks a drive alert as acknowledged (deactivates it)
func (rg *RouteGroup) AcknowledgeAlert(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || !user.IsAdmin {
		return apiErrors.ErrInsufficientPermissions()
	}

	alertIDStr := ctx.Params("alertId")
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid alert ID")
	}

	err = rg.driveMonitor.AcknowledgeAlert(ctx.Context(), alertID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to acknowledge alert")
	}

	return ctx.JSON(map[string]string{"message": "Alert acknowledged successfully"})
}