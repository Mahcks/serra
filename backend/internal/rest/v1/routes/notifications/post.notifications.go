package notifications

import (
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/websocket"
	"github.com/mahcks/serra/pkg/structures"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

// CreateNotification creates a new notification (admin only)
func (rg *RouteGroup) CreateNotification(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has admin permissions
	hasPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
		UserID:       user.ID,
		PermissionID: "admin.users",
	})
	if err != nil || !hasPermission {
		return apiErrors.ErrForbidden().SetDetail("Admin permission required")
	}

	var req structures.CreateNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid request body")
	}

	// Validate required fields
	if req.UserID == "" || req.Title == "" || req.Message == "" || req.Type == "" {
		return apiErrors.ErrBadRequest().SetDetail("Missing required fields")
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = structures.NotificationPriorityNormal
	}

	// Generate notification ID
	notificationID := uuid.New().String()

	// Prepare data for database
	var dataStr sql.NullString
	if req.Data != nil {
		if dataJson, err := req.Data.Value(); err == nil && dataJson != nil {
			dataStr = sql.NullString{
				String: string(dataJson.([]byte)),
				Valid:  true,
			}
		}
	}

	var expiresAt sql.NullTime
	if req.ExpiresAt != nil {
		expiresAt = sql.NullTime{
			Time:  *req.ExpiresAt,
			Valid: true,
		}
	}

	// Create notification in database
	err = rg.gctx.Crate().Sqlite.Query().CreateNotification(ctx.Context(), repository.CreateNotificationParams{
		ID:        notificationID,
		UserID:    req.UserID,
		Title:     req.Title,
		Message:   req.Message,
		Type:      string(req.Type),
		Priority:  string(req.Priority),
		Data:      dataStr,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("Failed to create notification", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create notification")
	}

	// Retrieve the created notification for response
	dbNotification, err := rg.gctx.Crate().Sqlite.Query().GetNotificationById(ctx.Context(), notificationID)
	if err != nil {
		slog.Error("Failed to retrieve created notification", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("Notification created but failed to retrieve")
	}

	// Convert to API response format
	notification := structures.Notification{
		ID:        dbNotification.ID,
		UserID:    dbNotification.UserID,
		Title:     dbNotification.Title,
		Message:   dbNotification.Message,
		Type:      structures.NotificationType(dbNotification.Type),
		Priority:  structures.NotificationPriority(dbNotification.Priority),
	}
	if dbNotification.CreatedAt.Valid {
		notification.CreatedAt = dbNotification.CreatedAt.Time
	}

	// Handle optional fields
	if dbNotification.Data.Valid {
		var data structures.NotificationData
		if err := data.Scan(dbNotification.Data.String); err == nil {
			notification.Data = &data
		}
	}

	if dbNotification.ReadAt.Valid {
		notification.ReadAt = &dbNotification.ReadAt.Time
	}

	if dbNotification.ExpiresAt.Valid {
		notification.ExpiresAt = &dbNotification.ExpiresAt.Time
	}

	// Broadcast notification via WebSocket
	payload := structures.NotificationWebSocketPayload{
		Type:         "notification_created",
		Notification: notification,
	}
	websocket.BroadcastToUser(req.UserID, structures.OpcodeNotification, payload)

	slog.Info("Notification created", "id", notificationID, "user_id", req.UserID, "type", req.Type)

	return ctx.Status(201).JSON(notification)
}