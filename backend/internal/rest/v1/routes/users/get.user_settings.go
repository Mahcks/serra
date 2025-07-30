package users

import (
	"database/sql"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetUserSettings retrieves comprehensive user settings including profile and preferences
func (rg *RouteGroup) GetUserSettings(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Get user profile information
	dbUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apiErrors.ErrNotFound().SetDetail("User not found")
		}
		slog.Error("Failed to get user profile", "error", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve user profile")
	}

	// Get user permissions
	permissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissions(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get user permissions", "error", err, "user_id", user.ID)
		// Don't fail the request, just log the error
		permissions = []repository.UserPermission{}
	}

	// Convert permissions to string slice
	permissionIDs := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permissionIDs = append(permissionIDs, perm.PermissionID)
	}

	// Get notification preferences
	notificationPrefs, err := rg.gctx.Crate().NotificationService.GetUserPreferences(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get notification preferences", "error", err, "user_id", user.ID)
		// Use defaults if we can't get preferences
		notificationPrefs = structures.DefaultNotificationPreferences()
		notificationPrefs.UserID = user.ID
	}

	// Build user settings response
	settings := structures.UserSettingsResponse{
		Profile: structures.UserProfile{
			ID:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     "",
			AvatarURL: "",
			UserType:  "local",
			CreatedAt: dbUser.CreatedAt.Time,
		},
		Permissions:            permissionIDs,
		NotificationPreferences: notificationPrefs,
		AccountSettings: structures.AccountSettings{
			Language:       "en",
			Theme:          "system",
			Timezone:       "UTC",
			DateFormat:     "YYYY-MM-DD",
			TimeFormat:     "24h",
		},
		PrivacySettings: structures.PrivacySettings{
			ShowOnlineStatus:    true,
			ShowWatchHistory:    false,
			ShowRequestHistory:  true,
		},
	}

	// Handle optional profile fields
	if dbUser.Email.Valid {
		settings.Profile.Email = dbUser.Email.String
	}
	if dbUser.AvatarUrl.Valid {
		settings.Profile.AvatarURL = dbUser.AvatarUrl.String
	}

	// Determine user type - check if user has media server integration
	// This would need to be determined based on your actual user model structure
	// For now, default to local unless we can determine otherwise
	settings.Profile.UserType = "local"

	return ctx.JSON(settings)
}