-- name: CreateMountedDrive :exec
INSERT INTO mounted_drives (
    id, name, mount_path, filesystem, total_size, used_size, 
    available_size, usage_percentage, is_online, last_checked
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetMountedDrive :one
SELECT * FROM mounted_drives WHERE id = ?;

-- name: GetMountedDriveByPath :one
SELECT * FROM mounted_drives WHERE mount_path = ?;

-- name: ListMountedDrives :many
SELECT * FROM mounted_drives ORDER BY name;

-- name: UpdateMountedDrive :exec
UPDATE mounted_drives SET
    name = ?,
    mount_path = ?,
    filesystem = ?,
    total_size = ?,
    used_size = ?,
    available_size = ?,
    usage_percentage = ?,
    is_online = ?,
    last_checked = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateMountedDriveStats :exec
UPDATE mounted_drives SET
    total_size = ?,
    used_size = ?,
    available_size = ?,
    usage_percentage = ?,
    is_online = ?,
    last_checked = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteMountedDrive :exec
DELETE FROM mounted_drives WHERE id = ?;

-- name: GetMountedDrivesForPolling :many
SELECT * FROM mounted_drives WHERE is_online = TRUE;