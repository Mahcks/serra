package notifications

import (
	"context"
	"log/slog"

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
	slog.Warn("GetUserPreferences temporarily returning hardcoded defaults - sqlc code not generated", "user_id", userID)
	
	// Return hardcoded defaults for now
	defaults := structures.DefaultNotificationPreferences()
	defaults.UserID = userID
	return defaults, nil
	
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
