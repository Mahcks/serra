package notifications

import (
	"database/sql"
	"log/slog"
	"strconv"
	"time"

	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

// GetNotifications retrieves paginated notifications for the authenticated user
func (rg *RouteGroup) GetNotifications(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Parse query parameters
	limitStr := ctx.Query("limit", "20")
	offsetStr := ctx.Query("offset", "0")
	unreadOnly := ctx.Query("unread_only", "false") == "true"
	notificationType := ctx.Query("type", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var dbNotifications []struct {
		ID        string
		UserID    string
		Title     string
		Message   string
		Type      string
		Priority  string
		Data      sql.NullString
		ReadAt    sql.NullTime
		CreatedAt time.Time
		ExpiresAt sql.NullTime
	}

	// Get notifications based on filters
	if unreadOnly {
		result, err := rg.gctx.Crate().Sqlite.Query().GetUnreadUserNotifications(ctx.Context(), user.ID)
		if err != nil {
			slog.Error("Failed to get unread notifications", "error", err, "user_id", user.ID)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve notifications")
		}

		// Convert to the expected format
		for _, n := range result {
			dbNotifications = append(dbNotifications, struct {
				ID        string
				UserID    string
				Title     string
				Message   string
				Type      string
				Priority  string
				Data      sql.NullString
				ReadAt    sql.NullTime
				CreatedAt time.Time
				ExpiresAt sql.NullTime
			}{
				ID:        n.ID,
				UserID:    n.UserID,
				Title:     n.Title,
				Message:   n.Message,
				Type:      n.Type,
				Priority:  n.Priority,
				Data:      n.Data,
				ReadAt:    n.ReadAt,
				CreatedAt: n.CreatedAt.Time,
				ExpiresAt: n.ExpiresAt,
			})
		}
	} else if notificationType != "" {
		result, err := rg.gctx.Crate().Sqlite.Query().GetNotificationsByType(ctx.Context(), repository.GetNotificationsByTypeParams{
			UserID: user.ID,
			Type:   notificationType,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			slog.Error("Failed to get notifications by type", "error", err, "user_id", user.ID, "type", notificationType)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve notifications")
		}

		// Convert to the expected format
		for _, n := range result {
			dbNotifications = append(dbNotifications, struct {
				ID        string
				UserID    string
				Title     string
				Message   string
				Type      string
				Priority  string
				Data      sql.NullString
				ReadAt    sql.NullTime
				CreatedAt time.Time
				ExpiresAt sql.NullTime
			}{
				ID:        n.ID,
				UserID:    n.UserID,
				Title:     n.Title,
				Message:   n.Message,
				Type:      n.Type,
				Priority:  n.Priority,
				Data:      n.Data,
				ReadAt:    n.ReadAt,
				CreatedAt: n.CreatedAt.Time,
				ExpiresAt: n.ExpiresAt,
			})
		}
	} else {
		result, err := rg.gctx.Crate().Sqlite.Query().GetUserNotifications(ctx.Context(), repository.GetUserNotificationsParams{
			UserID: user.ID,
			Limit:  int64(limit),
			Offset: int64(offset),
		})
		if err != nil {
			slog.Error("Failed to get user notifications", "error", err, "user_id", user.ID)
			return apiErrors.ErrInternalServerError().SetDetail("Failed to retrieve notifications")
		}

		// Convert to the expected format
		for _, n := range result {
			dbNotifications = append(dbNotifications, struct {
				ID        string
				UserID    string
				Title     string
				Message   string
				Type      string
				Priority  string
				Data      sql.NullString
				ReadAt    sql.NullTime
				CreatedAt time.Time
				ExpiresAt sql.NullTime
			}{
				ID:        n.ID,
				UserID:    n.UserID,
				Title:     n.Title,
				Message:   n.Message,
				Type:      n.Type,
				Priority:  n.Priority,
				Data:      n.Data,
				ReadAt:    n.ReadAt,
				CreatedAt: n.CreatedAt.Time,
				ExpiresAt: n.ExpiresAt,
			})
		}
	}

	// Get unread count
	unreadCount, err := rg.gctx.Crate().Sqlite.Query().CountUnreadNotifications(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to count unread notifications", "error", err, "user_id", user.ID)
		unreadCount = 0
	}

	// Convert to API response format
	notifications := make([]structures.Notification, 0, len(dbNotifications))
	for _, dbNotif := range dbNotifications {
		notification := structures.Notification{
			ID:        dbNotif.ID,
			UserID:    dbNotif.UserID,
			Title:     dbNotif.Title,
			Message:   dbNotif.Message,
			Type:      structures.NotificationType(dbNotif.Type),
			Priority:  structures.NotificationPriority(dbNotif.Priority),
			CreatedAt: dbNotif.CreatedAt,
		}

		// Handle optional fields
		if dbNotif.Data.Valid {
			var data structures.NotificationData
			if err := data.Scan(dbNotif.Data.String); err == nil {
				notification.Data = &data
			}
		}

		if dbNotif.ReadAt.Valid {
			notification.ReadAt = &dbNotif.ReadAt.Time
		}

		if dbNotif.ExpiresAt.Valid {
			notification.ExpiresAt = &dbNotif.ExpiresAt.Time
		}

		notifications = append(notifications, notification)
	}

	response := structures.NotificationsResponse{
		Notifications: notifications,
		Total:         int64(len(notifications)),
		Unread:        unreadCount,
		Page:          offset/limit + 1,
		Limit:         limit,
		HasMore:       len(notifications) == limit,
	}

	return ctx.JSON(response)
}

// GetUnreadCount returns the count of unread notifications for the authenticated user
func (rg *RouteGroup) GetUnreadCount(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	count, err := rg.gctx.Crate().Sqlite.Query().CountUnreadNotifications(ctx.Context(), user.ID)
	if err != nil {
		slog.Error("Failed to count unread notifications", "error", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to count notifications")
	}

	response := structures.NotificationCountResponse{
		UnreadCount: count,
	}

	return ctx.JSON(response)
}
