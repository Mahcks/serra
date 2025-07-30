package notifications

import (
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/websocket"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// DismissNotification deletes a notification
func (rg *RouteGroup) DismissNotification(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	notificationID := ctx.Params("id")
	if notificationID == "" {
		return apiErrors.ErrBadRequest().SetDetail("Notification ID is required")
	}

	// Verify notification exists and belongs to user before deletion
	notification, err := rg.gctx.Crate().Sqlite.Query().GetNotificationById(ctx.Context(), notificationID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("Notification not found")
	}

	if notification.UserID != user.ID {
		return apiErrors.ErrForbidden().SetDetail("Cannot access this notification")
	}

	// Delete the notification
	err = rg.gctx.Crate().Sqlite.Query().DismissNotification(ctx.Context(), repository.DismissNotificationParams{
		ID:     notificationID,
		UserID: user.ID,
	})
	if err != nil {
		slog.Error("Failed to dismiss notification", "error", err, "notification_id", notificationID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to dismiss notification")
	}

	// Convert to API response format for WebSocket
	notif := structures.Notification{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Title:     notification.Title,
		Message:   notification.Message,
		Type:      structures.NotificationType(notification.Type),
		Priority:  structures.NotificationPriority(notification.Priority),
	}
	if notification.CreatedAt.Valid {
		notif.CreatedAt = notification.CreatedAt.Time
	}

	// Handle optional fields
	if notification.Data.Valid {
		var data structures.NotificationData
		if err := data.Scan(notification.Data.String); err == nil {
			notif.Data = &data
		}
	}

	if notification.ReadAt.Valid {
		notif.ReadAt = &notification.ReadAt.Time
	}

	if notification.ExpiresAt.Valid {
		notif.ExpiresAt = &notification.ExpiresAt.Time
	}

	// Broadcast deletion via WebSocket
	payload := structures.NotificationWebSocketPayload{
		Type:         "notification_deleted",
		Notification: notif,
	}
	websocket.BroadcastToUser(user.ID, structures.OpcodeNotification, payload)

	slog.Info("Notification dismissed", "id", notificationID, "user_id", user.ID)

	return ctx.Status(204).Send(nil)
}

// DismissAllNotifications deletes all notifications for the user
func (rg *RouteGroup) DismissAllNotifications(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Delete all notifications for the user
	err := rg.gctx.Crate().Sqlite.Query().DismissAllUserNotifications(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to dismiss all notifications", "error", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to dismiss notifications")
	}

	response := map[string]interface{}{
		"message": "All notifications dismissed",
	}

	slog.Info("All notifications dismissed", "user_id", user.ID)

	return ctx.JSON(response)
}