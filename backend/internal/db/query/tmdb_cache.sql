-- name: GetCacheEntry :one
SELECT cache_key, data, endpoint, expires_at, created_at, updated_at
FROM tmdb_cache
WHERE cache_key = ? AND expires_at > CURRENT_TIMESTAMP;

-- name: SetCacheEntry :exec
INSERT OR REPLACE INTO tmdb_cache (cache_key, data, endpoint, expires_at, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP);

-- name: DeleteCacheEntry :exec
DELETE FROM tmdb_cache WHERE cache_key = ?;

-- name: DeleteExpiredCache :exec
DELETE FROM tmdb_cache WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: DeleteCacheByEndpoint :exec
DELETE FROM tmdb_cache WHERE endpoint = ?;

-- name: GetCacheStats :one
SELECT 
    COUNT(*) as total_entries,
    COUNT(CASE WHEN expires_at > CURRENT_TIMESTAMP THEN 1 END) as valid_entries,
    COUNT(CASE WHEN expires_at <= CURRENT_TIMESTAMP THEN 1 END) as expired_entries
FROM tmdb_cache;

-- name: GetCacheByEndpoint :many
SELECT cache_key, data, endpoint, expires_at, created_at, updated_at
FROM tmdb_cache
WHERE endpoint = ? AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- Static Data Queries
-- name: GetStaticData :one
SELECT data_type, data, last_updated
FROM tmdb_static_data
WHERE data_type = ?;

-- name: SetStaticData :exec
INSERT OR REPLACE INTO tmdb_static_data (data_type, data, last_updated)
VALUES (?, ?, CURRENT_TIMESTAMP);

-- name: GetAllStaticData :many
SELECT data_type, data, last_updated
FROM tmdb_static_data
ORDER BY data_type;

-- API Usage Tracking
-- name: IncrementAPIUsage :exec
INSERT INTO tmdb_api_usage (endpoint, request_count, date)
VALUES (?, 1, DATE('now'))
ON CONFLICT(endpoint, date) 
DO UPDATE SET request_count = request_count + 1;

-- name: GetAPIUsageToday :one
SELECT COALESCE(SUM(request_count), 0) as total_requests
FROM tmdb_api_usage
WHERE date = DATE('now');

-- name: GetAPIUsageByEndpoint :many
SELECT endpoint, SUM(request_count) as total_requests
FROM tmdb_api_usage
WHERE date >= DATE('now', '-7 days')
GROUP BY endpoint
ORDER BY total_requests DESC;

-- name: CleanupOldAPIUsage :exec
DELETE FROM tmdb_api_usage 
WHERE date < DATE('now', '-30 days');