package structures

import (
	"database/sql/driver"
	"time"
)

// NotificationPreferences represents user notification settings
type NotificationPreferences struct {
	ID                    string     `json:"id"`
	UserID                string     `json:"user_id"`
	RequestsApproved      bool       `json:"requests_approved"`
	RequestsDenied        bool       `json:"requests_denied"`
	DownloadCompleted     bool       `json:"download_completed"`
	MediaAvailable        bool       `json:"media_available"`
	SystemAlerts          bool       `json:"system_alerts"`
	MinPriority           string     `json:"min_priority"`
	WebNotifications      bool       `json:"web_notifications"`
	EmailNotifications    bool       `json:"email_notifications"`
	PushNotifications     bool       `json:"push_notifications"`
	QuietHoursEnabled     bool       `json:"quiet_hours_enabled"`
	QuietHoursStart       *string    `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd         *string    `json:"quiet_hours_end,omitempty"`
	AutoMarkReadAfterDays *int       `json:"auto_mark_read_after_days,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// UpdateNotificationPreferencesRequest represents the request to update user notification preferences
type UpdateNotificationPreferencesRequest struct {
	RequestsApproved      *bool   `json:"requests_approved,omitempty"`
	RequestsDenied        *bool   `json:"requests_denied,omitempty"`
	DownloadCompleted     *bool   `json:"download_completed,omitempty"`
	MediaAvailable        *bool   `json:"media_available,omitempty"`
	SystemAlerts          *bool   `json:"system_alerts,omitempty"`
	MinPriority           *string `json:"min_priority,omitempty"`
	WebNotifications      *bool   `json:"web_notifications,omitempty"`
	EmailNotifications    *bool   `json:"email_notifications,omitempty"`
	PushNotifications     *bool   `json:"push_notifications,omitempty"`
	QuietHoursEnabled     *bool   `json:"quiet_hours_enabled,omitempty"`
	QuietHoursStart       *string `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd         *string `json:"quiet_hours_end,omitempty"`
	AutoMarkReadAfterDays *int    `json:"auto_mark_read_after_days,omitempty"`
}

// NotificationPreferencesResponse represents the API response for notification preferences
type NotificationPreferencesResponse struct {
	Preferences         NotificationPreferences `json:"preferences"`
	AvailableTypes      []string               `json:"available_types"`
	AvailablePriorities []string               `json:"available_priorities"`
}

// NotificationPreferenceSummary represents a summary of user preferences for checking
type NotificationPreferenceSummary struct {
	Enabled           bool
	WebNotifications  bool
	MinPriority       string
	QuietHoursEnabled bool
	QuietHoursStart   *string
	QuietHoursEnd     *string
}

// IsNotificationAllowed checks if a notification type and priority should be sent to the user
func (p *NotificationPreferenceSummary) IsNotificationAllowed(notificationType string, priority NotificationPriority, currentTime time.Time) bool {
	// Check if notifications are enabled
	if !p.Enabled || !p.WebNotifications {
		return false
	}

	// Check priority level
	if !isPriorityAllowed(priority, p.MinPriority) {
		return false
	}

	// Check quiet hours
	if p.QuietHoursEnabled && p.isInQuietHours(currentTime) {
		// Only allow urgent notifications during quiet hours
		return priority == NotificationPriorityUrgent
	}

	return true
}

// isInQuietHours checks if the current time is within quiet hours
func (p *NotificationPreferenceSummary) isInQuietHours(currentTime time.Time) bool {
	if !p.QuietHoursEnabled || p.QuietHoursStart == nil || p.QuietHoursEnd == nil {
		return false
	}

	startTime, err := time.Parse("15:04", *p.QuietHoursStart)
	if err != nil {
		return false
	}

	endTime, err := time.Parse("15:04", *p.QuietHoursEnd)
	if err != nil {
		return false
	}

	currentTimeOfDay := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, time.UTC)
	startTimeOfDay := time.Date(0, 1, 1, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
	endTimeOfDay := time.Date(0, 1, 1, endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

	// Handle quiet hours that span midnight
	if startTimeOfDay.After(endTimeOfDay) {
		return currentTimeOfDay.After(startTimeOfDay) || currentTimeOfDay.Before(endTimeOfDay)
	}

	return currentTimeOfDay.After(startTimeOfDay) && currentTimeOfDay.Before(endTimeOfDay)
}

// isPriorityAllowed checks if a notification priority meets the minimum requirement
func isPriorityAllowed(notificationPriority NotificationPriority, minPriority string) bool {
	priorityLevels := map[string]int{
		"low":    1,
		"normal": 2,
		"high":   3,
		"urgent": 4,
	}

	notificationLevel, exists := priorityLevels[string(notificationPriority)]
	if !exists {
		return false
	}

	minLevel, exists := priorityLevels[minPriority]
	if !exists {
		return true // Default to allowing if min priority is invalid
	}

	return notificationLevel >= minLevel
}

// GetNotificationTypeForPreference maps notification types to preference fields
func GetNotificationTypeForPreference(notificationType NotificationType) string {
	switch notificationType {
	case NotificationTypeRequestApproved:
		return "request_approved"
	case NotificationTypeRequestDenied:
		return "request_denied"
	case NotificationTypeDownloadCompleted:
		return "download_completed"
	case NotificationTypeSystemAlert:
		return "system_alert"
	default:
		return "media_available"
	}
}

// DefaultNotificationPreferences returns default preferences for new users
func DefaultNotificationPreferences() NotificationPreferences {
	return NotificationPreferences{
		RequestsApproved:      true,
		RequestsDenied:        true,
		DownloadCompleted:     true,
		MediaAvailable:        true,
		SystemAlerts:          true,
		MinPriority:           "low",
		WebNotifications:      true,
		EmailNotifications:    false,
		PushNotifications:     false,
		QuietHoursEnabled:     false,
		AutoMarkReadAfterDays: nil,
	}
}

// Implement driver.Valuer interface for database storage
func (p NotificationPreferences) Value() (driver.Value, error) {
	// This struct is stored as individual columns, not JSON
	return nil, nil
}