package notifications

import (
	"context"
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

// CreateDefaultPreferencesForUser creates default notification preferences for a new user
func (s *Service) CreateDefaultPreferencesForUser(ctx context.Context, userID string) error {
	slog.Warn("CreateDefaultPreferencesForUser temporarily disabled - sqlc code not generated", "user_id", userID)
	return nil
	
	// TODO: Re-enable once sqlc generates the repository methods
	// prefsID := uuid.New().String()
	// err := s.query.CreateDefaultUserNotificationPreferences(ctx, prefsID, userID)
	// if err != nil {
	// 	slog.Error("Failed to create default notification preferences for user", "error", err, "user_id", userID)
	// 	return err
	// }
	// slog.Info("Created default notification preferences for user", "user_id", userID)
	// return nil
}

// EnsureUserHasPreferences ensures a user has notification preferences, creating defaults if needed
func (s *Service) EnsureUserHasPreferences(ctx context.Context, userID string) error {
	slog.Warn("EnsureUserHasPreferences temporarily disabled - sqlc code not generated", "user_id", userID)
	return nil
	
	// TODO: Re-enable once sqlc generates the repository methods
	// Check if preferences already exist
	// _, err := s.query.GetUserNotificationPreferences(ctx, userID)
	// if err == nil {
	// 	// Preferences already exist
	// 	return nil
	// }
	// Create default preferences
	// return s.CreateDefaultPreferencesForUser(ctx, userID)
}

// GetUserPreferences gets a user's notification preferences with defaults if they don't exist
func (s *Service) GetUserPreferences(ctx context.Context, userID string) (structures.NotificationPreferences, error) {
	// Get all user settings
	settings, err := s.query.GetAllUserSettings(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user notification preferences from settings", "error", err, "user_id", userID)
		// Return defaults on error
		defaults := structures.DefaultNotificationPreferences()
		defaults.UserID = userID
		return defaults, nil
	}

	// Convert settings to map for easier lookup
	settingsMap := make(map[string]string)
	for _, setting := range settings {
		settingsMap[setting.Key] = setting.Value
	}

	// Helper function to get boolean setting with default
	getBoolSetting := func(key string, defaultValue bool) bool {
		if value, exists := settingsMap[key]; exists {
			return value == "true"
		}
		return defaultValue
	}

	// Build preferences from user settings
	prefs := structures.NotificationPreferences{
		ID:                 "", // We don't use IDs in the simplified system
		UserID:             userID,
		RequestsApproved:   getBoolSetting("notifications_requests_approved", true),
		RequestsDenied:     getBoolSetting("notifications_requests_denied", true),
		DownloadCompleted:  getBoolSetting("notifications_download_completed", true),
		MediaAvailable:     getBoolSetting("notifications_media_available", true),
		SystemAlerts:       getBoolSetting("notifications_system_alerts", true),
		WebNotifications:   getBoolSetting("notifications_web_notifications", true),
		EmailNotifications: false, // Not implemented yet
		PushNotifications:  false, // Not implemented yet
	}

	return prefs, nil
	
	// TODO: Re-enable once sqlc generates the repository methods
	// Try to get existing preferences
	// dbPrefs, err := s.query.GetUserNotificationPreferences(ctx, userID)
	// if err != nil {
	// 	// If no preferences exist, create defaults
	// 	if err := s.CreateDefaultPreferencesForUser(ctx, userID); err != nil {
	// 		slog.Error("Failed to create default preferences", "error", err, "user_id", userID)
	// 		// Return hardcoded defaults if database operation fails
	// 		defaults := structures.DefaultNotificationPreferences()
	// 		defaults.UserID = userID
	// 		return defaults, nil
	// 	}

	// 	// Try to get preferences again
	// 	dbPrefs, err = s.query.GetUserNotificationPreferences(ctx, userID)
	// 	if err != nil {
	// 		slog.Error("Failed to get preferences after creation", "error", err, "user_id", userID)
	// 		// Return hardcoded defaults
	// 		defaults := structures.DefaultNotificationPreferences()
	// 		defaults.UserID = userID
	// 		return defaults, nil
	// 	}
	// }

	// // Convert to API format
	// prefs := structures.NotificationPreferences{
	// 	ID:                 dbPrefs.ID,
	// 	UserID:             dbPrefs.UserID,
	// 	RequestsApproved:   dbPrefs.RequestsApproved,
	// 	RequestsDenied:     dbPrefs.RequestsDenied,
	// 	DownloadCompleted:  dbPrefs.DownloadCompleted,
	// 	MediaAvailable:     dbPrefs.MediaAvailable,
	// 	SystemAlerts:       dbPrefs.SystemAlerts,
	// 	MinPriority:        dbPrefs.MinPriority,
	// 	WebNotifications:   dbPrefs.WebNotifications,
	// 	EmailNotifications: dbPrefs.EmailNotifications,
	// 	PushNotifications:  dbPrefs.PushNotifications,
	// 	QuietHoursEnabled:  dbPrefs.QuietHoursEnabled,
	// 	CreatedAt:          dbPrefs.CreatedAt,
	// 	UpdatedAt:          dbPrefs.UpdatedAt,
	// }

	// // Handle optional fields
	// if dbPrefs.QuietHoursStart.Valid {
	// 	prefs.QuietHoursStart = &dbPrefs.QuietHoursStart.String
	// }
	// if dbPrefs.QuietHoursEnd.Valid {
	// 	prefs.QuietHoursEnd = &dbPrefs.QuietHoursEnd.String
	// }
	// if dbPrefs.AutoMarkReadAfterDays.Valid {
	// 	days := int(dbPrefs.AutoMarkReadAfterDays.Int64)
	// 	prefs.AutoMarkReadAfterDays = &days
	// }

	// return prefs, nil
}

// UpdateUserPreferences updates a user's notification preferences using the user_settings key-value store
func (s *Service) UpdateUserPreferences(ctx context.Context, userID string, req structures.UpdateNotificationPreferencesRequest) error {
	slog.Info("Updating user notification preferences", "user_id", userID)
	
	// Update each preference that was provided
	if req.RequestsApproved != nil {
		value := "false"
		if *req.RequestsApproved {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_requests_approved",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update requests_approved preference", "error", err, "user_id", userID)
			return err
		}
	}

	if req.RequestsDenied != nil {
		value := "false"
		if *req.RequestsDenied {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_requests_denied",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update requests_denied preference", "error", err, "user_id", userID)
			return err
		}
	}

	if req.DownloadCompleted != nil {
		value := "false"
		if *req.DownloadCompleted {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_download_completed",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update download_completed preference", "error", err, "user_id", userID)
			return err
		}
	}

	if req.MediaAvailable != nil {
		value := "false"
		if *req.MediaAvailable {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_media_available",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update media_available preference", "error", err, "user_id", userID)
			return err
		}
	}

	if req.SystemAlerts != nil {
		value := "false"
		if *req.SystemAlerts {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_system_alerts",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update system_alerts preference", "error", err, "user_id", userID)
			return err
		}
	}

	if req.WebNotifications != nil {
		value := "false"
		if *req.WebNotifications {
			value = "true"
		}
		err := s.query.SetUserSetting(ctx, repository.SetUserSettingParams{
			UserID: userID,
			Key:    "notifications_web_notifications",
			Value:  value,
		})
		if err != nil {
			slog.Error("Failed to update web_notifications preference", "error", err, "user_id", userID)
			return err
		}
	}

	slog.Info("Successfully updated user notification preferences", "user_id", userID)
	return nil
}

// ShouldSendNotification checks if a notification should be sent based on user preferences
func (s *Service) ShouldSendNotification(ctx context.Context, userID string, notificationType string) bool {
	// Get user preferences
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user preferences for notification check", "error", err, "user_id", userID)
		// Default to sending notification if we can't check preferences
		return true
	}

	// Check the specific notification type
	switch notificationType {
	case "request_approved":
		return prefs.RequestsApproved && prefs.WebNotifications
	case "request_denied":
		return prefs.RequestsDenied && prefs.WebNotifications
	case "download_completed":
		// This covers both download completed and media available notifications
		return (prefs.DownloadCompleted || prefs.MediaAvailable) && prefs.WebNotifications
	case "media_available":
		return prefs.MediaAvailable && prefs.WebNotifications
	case "system_alert":
		return prefs.SystemAlerts && prefs.WebNotifications
	default:
		// For unknown types, default to true if web notifications are enabled
		return prefs.WebNotifications
	}
}
