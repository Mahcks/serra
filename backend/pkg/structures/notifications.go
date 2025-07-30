package structures

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo              NotificationType = "info"
	NotificationTypeSuccess           NotificationType = "success"
	NotificationTypeWarning           NotificationType = "warning"
	NotificationTypeError             NotificationType = "error"
	NotificationTypeDownloadCompleted NotificationType = "download_completed"
	NotificationTypeRequestApproved   NotificationType = "request_approved"
	NotificationTypeRequestDenied     NotificationType = "request_denied"
	NotificationTypeSystemAlert       NotificationType = "system_alert"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// NotificationData represents additional context data for notifications
type NotificationData struct {
	RequestID    *string `json:"request_id,omitempty"`
	DownloadID   *string `json:"download_id,omitempty"`
	MediaTitle   *string `json:"media_title,omitempty"`
	MediaType    *string `json:"media_type,omitempty"`
	TMDBID       *int64  `json:"tmdb_id,omitempty"`
	ActionURL    *string `json:"action_url,omitempty"`
	ImageURL     *string `json:"image_url,omitempty"`
	UserID       *string `json:"user_id,omitempty"`
	ErrorCode    *string `json:"error_code,omitempty"`
	Additional   map[string]interface{} `json:"additional,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (nd NotificationData) Value() (driver.Value, error) {
	// Check if all fields are nil/empty
	if nd.RequestID == nil && nd.DownloadID == nil && nd.MediaTitle == nil && 
	   nd.MediaType == nil && nd.TMDBID == nil && nd.ActionURL == nil && 
	   nd.ImageURL == nil && nd.UserID == nil && nd.ErrorCode == nil && 
	   len(nd.Additional) == 0 {
		return nil, nil
	}
	return json.Marshal(nd)
}

// Scan implements the sql.Scanner interface for database retrieval
func (nd *NotificationData) Scan(value interface{}) error {
	if value == nil {
		*nd = NotificationData{}
		return nil
	}
	
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), nd)
	case []byte:
		return json.Unmarshal(v, nd)
	default:
		*nd = NotificationData{}
		return nil
	}
}

// Notification represents a user notification
type Notification struct {
	ID        string                `json:"id"`
	UserID    string                `json:"user_id"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Type      NotificationType      `json:"type"`
	Priority  NotificationPriority  `json:"priority"`
	Data      *NotificationData     `json:"data,omitempty"`
	ReadAt    *time.Time            `json:"read_at,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	ExpiresAt *time.Time            `json:"expires_at,omitempty"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID    string                `json:"user_id" binding:"required"`
	Title     string                `json:"title" binding:"required"`
	Message   string                `json:"message" binding:"required"`
	Type      NotificationType      `json:"type" binding:"required"`
	Priority  NotificationPriority  `json:"priority,omitempty"`
	Data      *NotificationData     `json:"data,omitempty"`
	ExpiresAt *time.Time            `json:"expires_at,omitempty"`
}

// NotificationsResponse represents a paginated list of notifications
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	Unread        int64          `json:"unread"`
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
	HasMore       bool           `json:"has_more"`
}

// NotificationCountResponse represents unread notification count
type NotificationCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}

// WebSocket notification payload
type NotificationWebSocketPayload struct {
	Type         string       `json:"type"` // "notification_created", "notification_updated", "notification_deleted"
	Notification Notification `json:"notification"`
}

// Predefined notification templates
func NewDownloadCompletedNotification(userID, title, mediaTitle string, data *NotificationData) CreateNotificationRequest {
	if data == nil {
		data = &NotificationData{}
	}
	data.MediaTitle = &mediaTitle
	
	return CreateNotificationRequest{
		UserID:   userID,
		Title:    "Download Completed",
		Message:  "Your download of \"" + mediaTitle + "\" has completed successfully",
		Type:     NotificationTypeDownloadCompleted,
		Priority: NotificationPriorityNormal,
		Data:     data,
	}
}

func NewRequestApprovedNotification(userID, mediaTitle string, data *NotificationData) CreateNotificationRequest {
	if data == nil {
		data = &NotificationData{}
	}
	data.MediaTitle = &mediaTitle
	
	return CreateNotificationRequest{
		UserID:   userID,
		Title:    "Request Approved",
		Message:  "Your request for \"" + mediaTitle + "\" has been approved and is now being processed",
		Type:     NotificationTypeRequestApproved,
		Priority: NotificationPriorityNormal,
		Data:     data,
	}
}

func NewRequestDeniedNotification(userID, mediaTitle, reason string, data *NotificationData) CreateNotificationRequest {
	if data == nil {
		data = &NotificationData{}
	}
	data.MediaTitle = &mediaTitle
	
	message := "Your request for \"" + mediaTitle + "\" has been denied"
	if reason != "" {
		message += ": " + reason
	}
	
	return CreateNotificationRequest{
		UserID:   userID,
		Title:    "Request Denied",
		Message:  message,
		Type:     NotificationTypeRequestDenied,
		Priority: NotificationPriorityNormal,
		Data:     data,
	}
}

func NewSystemAlertNotification(userID, title, message string, priority NotificationPriority, data *NotificationData) CreateNotificationRequest {
	return CreateNotificationRequest{
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     NotificationTypeSystemAlert,
		Priority: priority,
		Data:     data,
	}
}