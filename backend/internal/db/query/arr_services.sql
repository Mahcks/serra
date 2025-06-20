-- name: CreateArrService :exec
INSERT INTO arr_services (id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetArrServiceByType :many
SELECT id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability
FROM arr_services
WHERE type = :arrType;