-- name: GetUserSetting :one
SELECT value 
FROM user_settings 
WHERE user_id = ? AND key = ?;

-- name: GetAllUserSettings :many
SELECT key, value, updated_at
FROM user_settings 
WHERE user_id = ?
ORDER BY key ASC;

-- name: SetUserSetting :exec
INSERT INTO user_settings (user_id, key, value, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT (user_id, key) 
DO UPDATE SET 
    value = excluded.value,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteUserSetting :exec
DELETE FROM user_settings 
WHERE user_id = ? AND key = ?;

-- name: DeleteAllUserSettings :exec
DELETE FROM user_settings 
WHERE user_id = ?;