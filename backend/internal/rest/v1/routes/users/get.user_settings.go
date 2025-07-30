package users

import (
	"database/sql"
	"log/slog"
	"strings"

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

	// Get user permissions with details
	permissions, err := rg.gctx.Crate().Sqlite.Query().GetUserPermissionsWithDetails(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get user permissions", "error", err, "user_id", user.ID)
		// Don't fail the request, just log the error
		permissions = []repository.GetUserPermissionsWithDetailsRow{}
	}

	// Convert permissions to detailed format
	permissionDetails := make([]structures.PermissionInfo, 0, len(permissions))
	for _, perm := range permissions {
		// Identify dangerous/admin permissions
		isDangerous := perm.PermissionID == "owner" || 
			strings.HasPrefix(perm.PermissionID, "admin.") ||
			perm.PermissionID == "requests.manage"
		
		// Determine category based on permission ID
		var category string
		switch {
		case perm.PermissionID == "owner":
			category = "System"
		case strings.HasPrefix(perm.PermissionID, "admin."):
			category = "Administration"
		case strings.HasPrefix(perm.PermissionID, "request."):
			category = "Requests"
		case strings.HasPrefix(perm.PermissionID, "requests."):
			category = "Request Management"
		default:
			category = "General"
		}
		
		description := ""
		if perm.Description.Valid {
			description = perm.Description.String
		}
		
		permissionDetails = append(permissionDetails, structures.PermissionInfo{
			ID:          perm.PermissionID,
			Name:        perm.Name,
			Description: description,
			Category:    category,
			Dangerous:   isDangerous,
		})
	}

	// Get notification preferences
	notificationPrefs, err := rg.gctx.Crate().NotificationService.GetUserPreferences(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get notification preferences", "error", err, "user_id", user.ID)
		// Use defaults if we can't get preferences
		notificationPrefs = structures.DefaultNotificationPreferences()
		notificationPrefs.UserID = user.ID
	}

	// Get user settings from key-value store
	userSettings, err := rg.gctx.Crate().Sqlite.Query().GetAllUserSettings(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to get user settings", "error", err, "user_id", user.ID)
		// Don't fail the request, just use defaults if we can't get settings
		userSettings = []repository.GetAllUserSettingsRow{}
	}

	// Convert settings to map for easier lookup
	settingsMap := make(map[string]string)
	for _, setting := range userSettings {
		settingsMap[setting.Key] = setting.Value
	}

	// Helper function to get setting with default
	getSetting := func(key, defaultValue string) string {
		if value, exists := settingsMap[key]; exists {
			return value
		}
		return defaultValue
	}

	// Helper function to get boolean setting with default
	getBoolSetting := func(key string, defaultValue bool) bool {
		if value, exists := settingsMap[key]; exists {
			return value == "true"
		}
		return defaultValue
	}

	// Build user settings response
	settings := structures.UserSettingsResponse{
		Profile: structures.UserProfile{
			ID:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     "",
			AvatarURL: "",
			UserType:  dbUser.UserType,
			CreatedAt: dbUser.CreatedAt.Time,
		},
		Permissions:            permissionDetails,
		NotificationPreferences: notificationPrefs,
		AccountSettings: structures.AccountSettings{
			Language:       getSetting("language", "en"),
			Theme:          getSetting("theme", "system"),
			Timezone:       getSetting("timezone", "UTC"),
			DateFormat:     getSetting("date_format", "YYYY-MM-DD"),
			TimeFormat:     getSetting("time_format", "24h"),
		},
		PrivacySettings: structures.PrivacySettings{
			ShowOnlineStatus:    getBoolSetting("show_online_status", true),
			ShowWatchHistory:    getBoolSetting("show_watch_history", false),
			ShowRequestHistory:  getBoolSetting("show_request_history", true),
		},
	}

	// Handle optional profile fields
	if dbUser.Email.Valid {
		settings.Profile.Email = dbUser.Email.String
	}
	if dbUser.AvatarUrl.Valid {
		settings.Profile.AvatarURL = dbUser.AvatarUrl.String
	}

	// User type is already set from dbUser.UserType above

	return ctx.JSON(settings)
}