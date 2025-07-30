package notifications

import (
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetNotificationPreferences retrieves the user's notification preferences
func (rg *RouteGroup) GetNotificationPreferences(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// TODO: Re-enable once sqlc generates the method
	// For now, return default preferences
	slog.Warn("GetUserNotificationPreferences temporarily disabled - returning defaults", "user_id", user.ID)
	
	defaultPrefs := structures.DefaultNotificationPreferences()
	defaultPrefs.UserID = user.ID
	
	response := structures.NotificationPreferencesResponse{
		Preferences: defaultPrefs,
		AvailableTypes: []string{
			"requests_approved",
			"requests_denied",
			"download_completed", 
			"media_available",
			"system_alerts",
		},
		AvailablePriorities: []string{"low", "normal", "high", "urgent"},
	}
	
	return ctx.JSON(response)
}