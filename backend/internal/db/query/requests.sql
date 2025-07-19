-- name: CreateRequest :one
INSERT INTO requests (user_id, media_type, tmdb_id, title, status, notes, poster_url, on_behalf_of, seasons, season_statuses)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses;

-- name: GetRequestByID :one
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE id = ?;

-- name: GetRequestsByUser :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE user_id = ?
ORDER BY created_at DESC;

-- name: GetAllRequests :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
ORDER BY created_at DESC;

-- name: GetRequestsByStatus :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE status = ?
ORDER BY created_at DESC;

-- name: GetPendingRequests :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE status = 'pending'
ORDER BY created_at ASC;

-- name: UpdateRequestStatus :one
UPDATE requests
SET status = ?, approver_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses;

-- name: UpdateRequestStatusOnly :one
UPDATE requests
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses;

-- name: FulfillRequest :one
UPDATE requests
SET status = 'fulfilled', fulfilled_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses;

-- name: DeleteRequest :exec
DELETE FROM requests WHERE id = ?;

-- name: CheckExistingRequest :one
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE media_type = ? AND tmdb_id = ? AND user_id = ? AND seasons = ?;

-- name: CheckExistingRequestAnySeasons :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE media_type = ? AND tmdb_id = ? AND user_id = ?;

-- name: GetRequestsForUser :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE user_id = ? OR on_behalf_of = ?
ORDER BY created_at DESC;

-- name: GetRequestStatistics :one
SELECT 
    COUNT(*) as total_requests,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_requests,
    COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_requests,
    COUNT(CASE WHEN status = 'fulfilled' THEN 1 END) as fulfilled_requests,
    COUNT(CASE WHEN status = 'denied' THEN 1 END) as denied_requests
FROM requests;

-- name: GetRecentRequests :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE created_at >= datetime('now', '-7 days')
ORDER BY created_at DESC
LIMIT ?;

-- name: CheckMultipleUserRequests :many
SELECT tmdb_id, COUNT(*) > 0 as requested
FROM requests 
WHERE tmdb_id IN (/*SLICE:tmdb_ids*/?) AND media_type = ? AND user_id = ?
GROUP BY tmdb_id;

-- name: CheckUserRequestExists :one
SELECT COUNT(*) > 0 as requested
FROM requests 
WHERE tmdb_id = ? AND media_type = ? AND user_id = ?;

-- name: GetRequestsByTMDBIDAndMediaType :many
SELECT id, user_id, media_type, tmdb_id, title, status, notes, created_at, updated_at, fulfilled_at, approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests
WHERE tmdb_id = ? AND media_type = ?;