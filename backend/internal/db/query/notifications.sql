-- name: CreateNotification :exec
INSERT INTO notifications (
    id, user_id, title, message, type, priority, data, expires_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetUserNotifications :many
SELECT 
    id, user_id, title, message, type, priority, data, read_at, created_at, expires_at
FROM notifications 
WHERE user_id = ? 
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY 
    CASE priority 
        WHEN 'urgent' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'normal' THEN 3 
        WHEN 'low' THEN 4 
    END ASC,
    created_at DESC
LIMIT ? OFFSET ?;

-- name: GetUnreadUserNotifications :many
SELECT 
    id, user_id, title, message, type, priority, data, read_at, created_at, expires_at
FROM notifications 
WHERE user_id = ? 
    AND read_at IS NULL
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY 
    CASE priority 
        WHEN 'urgent' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'normal' THEN 3 
        WHEN 'low' THEN 4 
    END ASC,
    created_at DESC;

-- name: GetNotificationById :one
SELECT 
    id, user_id, title, message, type, priority, data, read_at, created_at, expires_at
FROM notifications 
WHERE id = ?;

-- name: MarkNotificationAsRead :exec
UPDATE notifications 
SET read_at = CURRENT_TIMESTAMP 
WHERE id = ? AND user_id = ?;

-- name: DismissNotification :exec
DELETE FROM notifications 
WHERE id = ? AND user_id = ?;

-- name: DismissAllUserNotifications :exec
DELETE FROM notifications 
WHERE user_id = ?;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) 
FROM notifications 
WHERE user_id = ? 
    AND read_at IS NULL
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP);

-- name: CleanupExpiredNotifications :exec
DELETE FROM notifications 
WHERE expires_at IS NOT NULL 
    AND expires_at <= CURRENT_TIMESTAMP;

-- name: GetNotificationsByType :many
SELECT 
    id, user_id, title, message, type, priority, data, read_at, created_at, expires_at
FROM notifications 
WHERE user_id = ? 
    AND type = ?
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: BulkMarkAsRead :exec
UPDATE notifications 
SET read_at = CURRENT_TIMESTAMP 
WHERE id IN (sqlc.slice('notification_ids'))
    AND user_id = ?;

-- name: GetRecentNotifications :many
SELECT 
    id, user_id, title, message, type, priority, data, read_at, created_at, expires_at
FROM notifications 
WHERE user_id = ? 
    AND created_at >= datetime('now', '-7 days')
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC
LIMIT ?;