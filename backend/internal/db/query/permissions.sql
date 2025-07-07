-- name: CreatePermission :one
INSERT INTO permissions (id, name, description)
VALUES (:id, :name, :description)
RETURNING *;

-- name: GetPermission :one
SELECT * FROM permissions
WHERE id = :id;

-- name: GetPermissionByName :one
SELECT * FROM permissions
WHERE name = :name;

-- name: GetAllPermissions :many
SELECT * FROM permissions
ORDER BY name;

-- name: UpdatePermission :one
UPDATE permissions
SET name = :name,
    description = :description
WHERE id = :id
RETURNING *;

-- name: DeletePermission :exec
DELETE FROM permissions
WHERE id = :id;

-- name: CheckPermissionExists :one
SELECT COUNT(*) > 0 as permission_exists
FROM permissions
WHERE id = :id;