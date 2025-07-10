-- name: CreateArrService :exec
INSERT INTO arr_services (id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability, is_4k)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetArrServiceByType :many
SELECT id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability, is_4k, created_at
FROM arr_services
WHERE type = ?;