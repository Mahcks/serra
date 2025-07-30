package notifications

import (
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// UpdateNotificationPreferences updates the user's notification preferences
func (rg *RouteGroup) UpdateNotificationPreferences(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	var req structures.UpdateNotificationPreferencesRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Validate priority if provided
	if req.MinPriority != nil {
		validPriorities := map[string]bool{
			"low": true, "normal": true, "high": true, "urgent": true,
		}
		if !validPriorities[*req.MinPriority] {
			return apiErrors.ErrBadRequest().SetDetail("Invalid min_priority. Must be one of: low, normal, high, urgent")
		}
	}

	// Validate quiet hours format if provided
	if req.QuietHoursStart != nil && *req.QuietHoursStart != "" {
		if len(*req.QuietHoursStart) != 5 {
			return apiErrors.ErrBadRequest().SetDetail("Invalid quiet_hours_start format. Use HH:MM (24-hour format)")
		}
	}
	if req.QuietHoursEnd != nil && *req.QuietHoursEnd != "" {
		if len(*req.QuietHoursEnd) != 5 {
			return apiErrors.ErrBadRequest().SetDetail("Invalid quiet_hours_end format. Use HH:MM (24-hour format)")
		}
	}

	// TODO: Re-enable once sqlc generates the method
	// For now, just return success without actually updating preferences
	slog.Warn("UpdateNotificationPreferences temporarily disabled - sqlc code not generated", "user_id", user.ID)
	
	return ctx.JSON(map[string]interface{}{
		"message": "Notification preferences update temporarily disabled",
		"success": true,
	})
}