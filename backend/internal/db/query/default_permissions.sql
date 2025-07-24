-- name: GetDefaultPermissions :many
SELECT permission_id, enabled FROM default_permissions WHERE enabled = TRUE;

-- name: GetAllDefaultPermissionSettings :many  
SELECT permission_id, enabled FROM default_permissions ORDER BY permission_id;

-- name: UpdateDefaultPermission :exec
INSERT OR REPLACE INTO default_permissions (permission_id, enabled, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP);

-- name: EnsureDefaultPermissionExists :exec
INSERT OR IGNORE INTO default_permissions (permission_id, enabled)
VALUES (?, FALSE);

-- name: RemoveDefaultPermission :exec
DELETE FROM default_permissions WHERE permission_id = ?;