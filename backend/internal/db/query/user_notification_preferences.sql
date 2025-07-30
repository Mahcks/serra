-- name: GetUserNotificationPreferences :one
SELECT 
    id, user_id, requests_approved, requests_denied, download_completed, media_available, system_alerts,
    min_priority, web_notifications, email_notifications, push_notifications,
    quiet_hours_enabled, quiet_hours_start, quiet_hours_end, auto_mark_read_after_days,
    created_at, updated_at
FROM user_notification_preferences 
WHERE user_id = ?;

-- name: CreateUserNotificationPreferences :exec
INSERT INTO user_notification_preferences (
    id, user_id, requests_approved, requests_denied, download_completed, media_available, system_alerts,
    min_priority, web_notifications, email_notifications, push_notifications,
    quiet_hours_enabled, quiet_hours_start, quiet_hours_end, auto_mark_read_after_days
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateUserNotificationPreferences :exec
UPDATE user_notification_preferences SET
    requests_approved = ?,
    requests_denied = ?,
    download_completed = ?,
    media_available = ?,
    system_alerts = ?,
    min_priority = ?,
    web_notifications = ?,
    email_notifications = ?,
    push_notifications = ?,
    quiet_hours_enabled = ?,
    quiet_hours_start = ?,
    quiet_hours_end = ?,
    auto_mark_read_after_days = ?
WHERE user_id = ?;

-- name: CreateDefaultUserNotificationPreferences :exec
INSERT INTO user_notification_preferences (id, user_id)
VALUES (?, ?)
ON CONFLICT(user_id) DO NOTHING;

-- name: GetUsersWithNotificationPreference :many
SELECT DISTINCT user_id
FROM user_notification_preferences
WHERE 
    CASE 
        WHEN ? = 'request_approved' THEN requests_approved = TRUE
        WHEN ? = 'request_denied' THEN requests_denied = TRUE
        WHEN ? = 'download_completed' THEN download_completed = TRUE
        WHEN ? = 'media_available' THEN media_available = TRUE
        WHEN ? = 'system_alert' THEN system_alerts = TRUE
        ELSE TRUE
    END
    AND web_notifications = TRUE;

-- name: CheckUserNotificationEnabled :one
SELECT 
    CASE 
        WHEN ? = 'request_approved' THEN requests_approved
        WHEN ? = 'request_denied' THEN requests_denied
        WHEN ? = 'download_completed' THEN download_completed
        WHEN ? = 'media_available' THEN media_available
        WHEN ? = 'system_alert' THEN system_alerts
        ELSE TRUE
    END as enabled,
    web_notifications,
    min_priority,
    quiet_hours_enabled,
    quiet_hours_start,
    quiet_hours_end
FROM user_notification_preferences 
WHERE user_id = ?;

-- name: GetUsersForQuietHoursCleanup :many
SELECT user_id, auto_mark_read_after_days
FROM user_notification_preferences
WHERE auto_mark_read_after_days IS NOT NULL
    AND auto_mark_read_after_days > 0;

-- name: DeleteUserNotificationPreferences :exec
DELETE FROM user_notification_preferences 
WHERE user_id = ?;