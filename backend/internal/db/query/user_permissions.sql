-- name: AssignUserPermission :exec
INSERT INTO user_permissions (user_id, permission_id)
VALUES (:user_id, :permission_id)
ON CONFLICT (user_id, permission_id) DO NOTHING;

-- name: RevokeUserPermission :exec
DELETE FROM user_permissions
WHERE user_id = :user_id AND permission_id = :permission_id;

-- name: GetUserPermissions :many
SELECT user_id, permission_id
FROM user_permissions
WHERE user_id = :user_id;

-- name: GetUserPermissionsWithDetails :many
SELECT up.user_id, up.permission_id, p.name, p.description
FROM user_permissions up
JOIN permissions p ON up.permission_id = p.id
WHERE up.user_id = :user_id;

-- name: CheckUserPermission :one
SELECT COUNT(*) > 0 as has_permission
FROM user_permissions
WHERE user_id = :user_id AND permission_id = :permission_id;

-- name: GetAllUserPermissions :many
SELECT up.user_id, up.permission_id, u.username, p.name, p.description
FROM user_permissions up
JOIN users u ON up.user_id = u.id
JOIN permissions p ON up.permission_id = p.id
ORDER BY u.username, p.name;

-- name: DeleteUserPermissions :exec
DELETE FROM user_permissions WHERE user_id = :user_id;