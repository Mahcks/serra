package notifications

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/pkg/structures"
)

// BroadcastFunc defines the function signature for broadcasting WebSocket messages
type BroadcastFunc func(userID string, op structures.Opcode, data interface{})

type Service struct {
	query     *repository.Queries
	broadcast BroadcastFunc
}

func NewService(query *repository.Queries) *Service {
	return &Service{
		query: query,
	}
}

// SetBroadcastFunc sets the WebSocket broadcast function
func (s *Service) SetBroadcastFunc(broadcast BroadcastFunc) {
	s.broadcast = broadcast
}

// CheckUserPreferences checks if a user should receive a notification based on their preferences
func (s *Service) CheckUserPreferences(ctx context.Context, userID string, notificationType structures.NotificationType, priority structures.NotificationPriority) (*structures.NotificationPreferenceSummary, bool) {
	// Convert notification type to string for preference checking
	var typeStr string
	switch notificationType {
	case structures.NotificationTypeRequestApproved:
		typeStr = "request_approved"
	case structures.NotificationTypeRequestDenied:
		typeStr = "request_denied"
	case structures.NotificationTypeDownloadCompleted:
		typeStr = "download_completed"
	case structures.NotificationTypeSystemAlert:
		typeStr = "system_alert"
	default:
		typeStr = string(notificationType)
	}

	// Use the ShouldSendNotification method
	allowed := s.ShouldSendNotification(ctx, userID, typeStr)
	
	// Get full preferences for summary
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get user preferences for summary", "error", err, "user_id", userID)
		// Return default summary but respect the allowed flag
		return &structures.NotificationPreferenceSummary{
			Enabled:          allowed,
			WebNotifications: allowed,
			MinPriority:      "low",
		}, allowed
	}

	// Create summary from preferences
	summary := &structures.NotificationPreferenceSummary{
		Enabled:          allowed,
		WebNotifications: prefs.WebNotifications,
		MinPriority:      "low", // We don't implement priority filtering yet
	}

	if !allowed {
		slog.Debug("Notification blocked by user preferences", 
			"user_id", userID, 
			"type", notificationType, 
			"priority", priority,
			"enabled", summary.Enabled,
			"web_notifications", summary.WebNotifications)
	}

	return summary, allowed
}

// CreateNotification creates a notification and broadcasts it via WebSocket
func (s *Service) CreateNotification(ctx context.Context, userID string, notification structures.CreateNotificationRequest) error {
	// Check user preferences first
	_, allowed := s.CheckUserPreferences(ctx, userID, notification.Type, notification.Priority)
	if !allowed {
		slog.Debug("Notification not sent due to user preferences", "user_id", userID, "type", notification.Type)
		return nil // Not an error, just filtered out
	}
	// Generate notification ID
	notificationID := uuid.New().String()

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = structures.NotificationPriorityNormal
	}

	// Prepare data for database
	var dataStr sql.NullString
	if notification.Data != nil {
		if dataJson, err := notification.Data.Value(); err == nil && dataJson != nil {
			dataStr = sql.NullString{
				String: string(dataJson.([]byte)),
				Valid:  true,
			}
		}
	}

	var expiresAt sql.NullTime
	if notification.ExpiresAt != nil {
		expiresAt = sql.NullTime{
			Time:  *notification.ExpiresAt,
			Valid: true,
		}
	}

	// Create notification in database
	err := s.query.CreateNotification(ctx, repository.CreateNotificationParams{
		ID:        notificationID,
		UserID:    userID,
		Title:     notification.Title,
		Message:   notification.Message,
		Type:      string(notification.Type),
		Priority:  string(notification.Priority),
		Data:      dataStr,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("Failed to create notification", "error", err, "user_id", userID)
		return err
	}

	// Retrieve the created notification for WebSocket broadcast
	dbNotification, err := s.query.GetNotificationById(ctx, notificationID)
	if err != nil {
		slog.Error("Failed to retrieve created notification", "error", err, "notification_id", notificationID)
		return err
	}

	// Convert to API response format
	notif := structures.Notification{
		ID:       dbNotification.ID,
		UserID:   dbNotification.UserID,
		Title:    dbNotification.Title,
		Message:  dbNotification.Message,
		Type:     structures.NotificationType(dbNotification.Type),
		Priority: structures.NotificationPriority(dbNotification.Priority),
	}

	if dbNotification.CreatedAt.Valid {
		notif.CreatedAt = dbNotification.CreatedAt.Time
	}

	// Handle optional fields
	if dbNotification.Data.Valid {
		var data structures.NotificationData
		if err := data.Scan(dbNotification.Data.String); err == nil {
			notif.Data = &data
		}
	}

	if dbNotification.ReadAt.Valid {
		notif.ReadAt = &dbNotification.ReadAt.Time
	}

	if dbNotification.ExpiresAt.Valid {
		notif.ExpiresAt = &dbNotification.ExpiresAt.Time
	}

	// Broadcast notification via WebSocket
	payload := structures.NotificationWebSocketPayload{
		Type:         "notification_created",
		Notification: notif,
	}
	if s.broadcast != nil {
		s.broadcast(userID, structures.OpcodeNotification, payload)
	}

	slog.Info("Notification created", "id", notificationID, "user_id", userID, "type", notification.Type)
	return nil
}

// NotifyMediaAvailable notifies a user that their requested media is now available
func (s *Service) NotifyMediaAvailable(ctx context.Context, userID string, mediaTitle, mediaType string, tmdbID *int64) error {
	data := &structures.NotificationData{
		MediaTitle: &mediaTitle,
		MediaType:  &mediaType,
		TMDBID:     tmdbID,
	}

	notification := structures.CreateNotificationRequest{
		UserID:   userID,
		Title:    "Media Available",
		Message:  mediaTitle + " is now available for streaming!",
		Type:     structures.NotificationTypeDownloadCompleted,
		Priority: structures.NotificationPriorityHigh,
		Data:     data,
	}

	return s.CreateNotification(ctx, userID, notification)
}

// NotifyRequestApproved notifies a user that their media request was approved
func (s *Service) NotifyRequestApproved(ctx context.Context, userID string, mediaTitle, mediaType string, tmdbID *int64, requestID *string) error {
	data := &structures.NotificationData{
		MediaTitle: &mediaTitle,
		MediaType:  &mediaType,
		TMDBID:     tmdbID,
		RequestID:  requestID,
	}

	notification := structures.CreateNotificationRequest{
		UserID:   userID,
		Title:    "Request Approved",
		Message:  "Your request for " + mediaTitle + " has been approved and is being processed.",
		Type:     structures.NotificationTypeRequestApproved,
		Priority: structures.NotificationPriorityNormal,
		Data:     data,
	}

	return s.CreateNotification(ctx, userID, notification)
}

// NotifyRequestDenied notifies a user that their media request was denied
func (s *Service) NotifyRequestDenied(ctx context.Context, userID string, mediaTitle, mediaType, reason string, tmdbID *int64, requestID *string) error {
	data := &structures.NotificationData{
		MediaTitle: &mediaTitle,
		MediaType:  &mediaType,
		TMDBID:     tmdbID,
		RequestID:  requestID,
	}

	message := "Your request for " + mediaTitle + " has been denied."
	if reason != "" {
		message += " Reason: " + reason
	}

	notification := structures.CreateNotificationRequest{
		UserID:   userID,
		Title:    "Request Denied",
		Message:  message,
		Type:     structures.NotificationTypeRequestDenied,
		Priority: structures.NotificationPriorityNormal,
		Data:     data,
	}

	return s.CreateNotification(ctx, userID, notification)
}

// NotifyDownloadCompleted notifies a user that a download has completed
func (s *Service) NotifyDownloadCompleted(ctx context.Context, userID string, mediaTitle, mediaType string, tmdbID *int64, downloadID *string) error {
	data := &structures.NotificationData{
		MediaTitle: &mediaTitle,
		MediaType:  &mediaType,
		TMDBID:     tmdbID,
		DownloadID: downloadID,
	}

	notification := structures.CreateNotificationRequest{
		UserID:   userID,
		Title:    "Download Complete",
		Message:  mediaTitle + " has finished downloading and is now available.",
		Type:     structures.NotificationTypeDownloadCompleted,
		Priority: structures.NotificationPriorityHigh,
		Data:     data,
	}

	return s.CreateNotification(ctx, userID, notification)
}

// NotifySystemAlert sends a system-wide alert to all users with admin permissions
func (s *Service) NotifySystemAlert(ctx context.Context, title, message string, priority structures.NotificationPriority) error {
	// Get all users with admin permissions
	adminUsers, err := s.query.GetAllUserPermissions(ctx)
	if err != nil {
		slog.Error("Failed to get admin users for system alert", "error", err)
		return err
	}

	// Filter for admin users
	var adminUserIDs []string
	for _, userPerm := range adminUsers {
		if userPerm.PermissionID == "admin.system" || userPerm.PermissionID == "owner" {
			adminUserIDs = append(adminUserIDs, userPerm.UserID)
		}
	}

	// Send notification to each admin user
	for _, userID := range adminUserIDs {
		notification := structures.CreateNotificationRequest{
			UserID:   userID,
			Title:    title,
			Message:  message,
			Type:     structures.NotificationTypeSystemAlert,
			Priority: priority,
		}

		if err := s.CreateNotification(ctx, userID, notification); err != nil {
			slog.Error("Failed to send system alert to user", "error", err, "user_id", userID)
		}
	}

	slog.Info("System alert sent", "title", title, "admin_count", len(adminUserIDs))
	return nil
}

// NotifyUserInvited notifies a user about a new invitation
func (s *Service) NotifyUserInvited(ctx context.Context, userID, inviterName, inviteCode string, expiresAt *time.Time) error {
	data := &structures.NotificationData{
		Additional: map[string]interface{}{
			"invite_code":   inviteCode,
			"inviter_name":  inviterName,
		},
	}

	message := "You've been invited to join Serra"
	if inviterName != "" {
		message += " by " + inviterName
	}

	notification := structures.CreateNotificationRequest{
		UserID:    userID,
		Title:     "Invitation Received",
		Message:   message,
		Type:      structures.NotificationTypeInfo,
		Priority:  structures.NotificationPriorityNormal,
		Data:      data,
		ExpiresAt: expiresAt,
	}

	return s.CreateNotification(ctx, userID, notification)
}

// NotifyBulkToUsers sends the same notification to multiple users
func (s *Service) NotifyBulkToUsers(ctx context.Context, userIDs []string, notification structures.CreateNotificationRequest) error {
	for _, userID := range userIDs {
		if err := s.CreateNotification(ctx, userID, notification); err != nil {
			slog.Error("Failed to send bulk notification to user", "error", err, "user_id", userID)
		}
	}

	slog.Info("Bulk notification sent", "title", notification.Title, "user_count", len(userIDs))
	return nil
}

// CleanupExpiredNotifications removes expired notifications
func (s *Service) CleanupExpiredNotifications(ctx context.Context) error {
	err := s.query.CleanupExpiredNotifications(ctx)
	if err != nil {
		slog.Error("Failed to cleanup expired notifications", "error", err)
		return err
	}

	slog.Info("Cleaned up expired notifications")
	return nil
}