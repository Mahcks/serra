package structures

import "time"

// UserProfile represents user profile information
type UserProfile struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	UserType  string    `json:"user_type"` // "local" or "media_server"
	CreatedAt time.Time `json:"created_at"`
}

// AccountSettings represents user account preferences
type AccountSettings struct {
	Language   string `json:"language"`   // "en", "es", "fr", etc.
	Theme      string `json:"theme"`      // "light", "dark", "system"
	Timezone   string `json:"timezone"`   // "UTC", "America/New_York", etc.
	DateFormat string `json:"date_format"` // "YYYY-MM-DD", "MM/DD/YYYY", etc.
	TimeFormat string `json:"time_format"` // "12h", "24h"
}

// PrivacySettings represents user privacy preferences
type PrivacySettings struct {
	ShowOnlineStatus   bool `json:"show_online_status"`
	ShowWatchHistory   bool `json:"show_watch_history"`
	ShowRequestHistory bool `json:"show_request_history"`
}

// UserSettingsResponse represents the complete user settings
type UserSettingsResponse struct {
	Profile                 UserProfile             `json:"profile"`
	Permissions             []PermissionInfo        `json:"permissions"`
	NotificationPreferences NotificationPreferences `json:"notification_preferences"`
	AccountSettings         AccountSettings         `json:"account_settings"`
	PrivacySettings         PrivacySettings         `json:"privacy_settings"`
}

// UpdateUserSettingsRequest represents a request to update user settings
type UpdateUserSettingsRequest struct {
	Profile                 *UpdateUserProfileRequest             `json:"profile,omitempty"`
	NotificationPreferences *UpdateNotificationPreferencesRequest `json:"notification_preferences,omitempty"`
	AccountSettings         *UpdateAccountSettingsRequest         `json:"account_settings,omitempty"`
	PrivacySettings         *UpdatePrivacySettingsRequest         `json:"privacy_settings,omitempty"`
}

// UpdateUserProfileRequest represents profile update fields
type UpdateUserProfileRequest struct {
	Email     *string `json:"email,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UpdateAccountSettingsRequest represents account settings update fields
type UpdateAccountSettingsRequest struct {
	Language   *string `json:"language,omitempty"`
	Theme      *string `json:"theme,omitempty"`
	Timezone   *string `json:"timezone,omitempty"`
	DateFormat *string `json:"date_format,omitempty"`
	TimeFormat *string `json:"time_format,omitempty"`
}

// UpdatePrivacySettingsRequest represents privacy settings update fields
type UpdatePrivacySettingsRequest struct {
	ShowOnlineStatus   *bool `json:"show_online_status,omitempty"`
	ShowWatchHistory   *bool `json:"show_watch_history,omitempty"`
	ShowRequestHistory *bool `json:"show_request_history,omitempty"`
}

// DeleteAccountRequest represents an account deletion request
type DeleteAccountRequest struct {
	Password      string `json:"password" binding:"required"`
	Confirmation  string `json:"confirmation" binding:"required"` // Must be "DELETE"
	Reason        string `json:"reason,omitempty"`
}