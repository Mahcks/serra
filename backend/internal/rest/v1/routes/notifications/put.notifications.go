package notifications

import (
	"log/slog"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/websocket"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// MarkAsRead marks a notification as read
func (rg *RouteGroup) MarkAsRead(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	notificationID := ctx.Params("id")
	if notificationID == "" {
		return apiErrors.ErrBadRequest().SetDetail("Notification ID is required")
	}

	// Verify notification exists and belongs to user
	notification, err := rg.gctx.Crate().Sqlite.Query().GetNotificationById(ctx.Context(), notificationID)
	if err != nil {
		return apiErrors.ErrNotFound().SetDetail("Notification not found")
	}

	if notification.UserID != user.ID {
		return apiErrors.ErrForbidden().SetDetail("Cannot access this notification")
	}

	// Mark as read
	err = rg.gctx.Crate().Sqlite.Query().MarkNotificationAsRead(ctx.Context(), repository.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: user.ID,
	})
	if err != nil {
		slog.Error("Failed to mark notification as read", "error", err, "notification_id", notificationID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to mark notification as read")
	}

	// Get updated notification for response
	updatedNotification, err := rg.gctx.Crate().Sqlite.Query().GetNotificationById(ctx.Context(), notificationID)
	if err != nil {
		slog.Error("Failed to retrieve updated notification", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Notification updated but failed to retrieve")
	}

	// Convert to API response format
	notif := structures.Notification{
		ID:        updatedNotification.ID,
		UserID:    updatedNotification.UserID,
		Title:     updatedNotification.Title,
		Message:   updatedNotification.Message,
		Type:      structures.NotificationType(updatedNotification.Type),
		Priority:  structures.NotificationPriority(updatedNotification.Priority),
	}
	if updatedNotification.CreatedAt.Valid {
		notif.CreatedAt = updatedNotification.CreatedAt.Time
	}

	// Handle optional fields
	if updatedNotification.Data.Valid {
		var data structures.NotificationData
		if err := data.Scan(updatedNotification.Data.String); err == nil {
			notif.Data = &data
		}
	}

	if updatedNotification.ReadAt.Valid {
		notif.ReadAt = &updatedNotification.ReadAt.Time
	}

	if updatedNotification.ExpiresAt.Valid {
		notif.ExpiresAt = &updatedNotification.ExpiresAt.Time
	}

	// Broadcast update via WebSocket
	payload := structures.NotificationWebSocketPayload{
		Type:         "notification_updated",
		Notification: notif,
	}
	websocket.BroadcastToUser(user.ID, structures.OpcodeNotification, payload)

	slog.Info("Notification marked as read", "id", notificationID, "user_id", user.ID)

	return ctx.JSON(notif)
}

// BulkMarkAsRead marks multiple notifications as read
func (rg *RouteGroup) BulkMarkAsRead(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	var req struct {
		NotificationIDs []string `json:"notification_ids" binding:"required"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	if len(req.NotificationIDs) == 0 {
		return apiErrors.ErrBadRequest().SetDetail("No notification IDs provided")
	}

	if len(req.NotificationIDs) > 50 {
		return apiErrors.ErrBadRequest().SetDetail("Too many notifications (max 50)")
	}

	// Mark notifications as read
	err := rg.gctx.Crate().Sqlite.Query().BulkMarkAsRead(ctx.Context(), repository.BulkMarkAsReadParams{
		NotificationIds: req.NotificationIDs,
		UserID:          user.ID,
	})
	if err != nil {
		slog.Error("Failed to bulk mark notifications as read", "error", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to mark notifications as read")
	}

	// Get updated count
	unreadCount, err := rg.gctx.Crate().Sqlite.Query().CountUnreadNotifications(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to count unread notifications", "error", err, "user_id", user.ID)
		unreadCount = 0
	}

	response := map[string]interface{}{
		"updated_count": len(req.NotificationIDs),
		"unread_count":  unreadCount,
	}

	slog.Info("Bulk marked notifications as read", "count", len(req.NotificationIDs), "user_id", user.ID)

	return ctx.JSON(response)
}