-- Request/Fulfillment Analytics Queries

-- GetRequestSuccessRates returns success rates by status and time period
-- name: GetRequestSuccessRates :many
SELECT 
    status,
    COUNT(*) as total_requests,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM requests 
WHERE created_at >= ?
GROUP BY status
ORDER BY COUNT(*) DESC;

-- GetRequestFulfillmentByUser returns request fulfillment metrics per user
-- name: GetRequestFulfillmentByUser :many
SELECT 
    u.username,
    COUNT(*) as total_requests,
    COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_requests,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_requests,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_requests,
    ROUND(COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM requests r
JOIN users u ON r.user_id = u.id
WHERE r.created_at >= ?
GROUP BY u.id, u.username
ORDER BY COUNT(*) DESC
LIMIT ?;

-- GetRequestTrends returns request trends over time periods
-- name: GetRequestTrends :many
SELECT 
    DATE(created_at) as request_date,
    COUNT(*) as total_requests,
    COUNT(CASE WHEN status = 'approved' THEN 1 END) as approved_requests,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_requests,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_requests,
    COUNT(CASE WHEN media_type = 'movie' THEN 1 END) as movie_requests,
    COUNT(CASE WHEN media_type = 'tv' THEN 1 END) as tv_requests
FROM requests 
WHERE created_at >= ?
GROUP BY DATE(created_at)
ORDER BY request_date DESC
LIMIT ?;

-- GetPopularRequestedContent returns most requested content with fulfillment status
-- name: GetPopularRequestedContent :many
SELECT 
    tmdb_id,
    title,
    media_type,
    COUNT(*) as request_count,
    COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) as fulfilled_count,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
    ROUND(COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) * 100.0 / COUNT(*), 2) as fulfillment_rate,
    MIN(created_at) as first_requested,
    MAX(created_at) as last_requested
FROM requests 
WHERE created_at >= ?
GROUP BY tmdb_id, title, media_type
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC, COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) DESC
LIMIT ?;

-- GetRequestProcessingPerformance returns processing performance metrics (simplified)
-- name: GetRequestProcessingPerformance :many
SELECT 
    media_type,
    status,
    COUNT(*) as count
FROM requests 
WHERE created_at >= ?
GROUP BY media_type, status
ORDER BY media_type, count DESC;

-- GetFailureAnalysis returns analysis of failed requests
-- name: GetFailureAnalysis :many
SELECT 
    media_type,
    COUNT(*) as total_failures,
    -- Extract common failure reasons from notes if available
    COUNT(CASE WHEN notes LIKE '%not found%' OR notes LIKE '%404%' THEN 1 END) as not_found_failures,
    COUNT(CASE WHEN notes LIKE '%timeout%' OR notes LIKE '%connection%' THEN 1 END) as connection_failures,
    COUNT(CASE WHEN notes LIKE '%quality%' OR notes LIKE '%profile%' THEN 1 END) as quality_failures,
    COUNT(CASE WHEN notes LIKE '%space%' OR notes LIKE '%disk%' THEN 1 END) as storage_failures
FROM requests 
WHERE status = 'failed' 
AND created_at >= ?
GROUP BY media_type
ORDER BY COUNT(*) DESC;

-- GetRequestVolumeByHour returns request volume patterns by hour of day
-- name: GetRequestVolumeByHour :many
SELECT 
    CAST(strftime('%H', created_at) AS INTEGER) as hour_of_day,
    COUNT(*) as total_requests,
    COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) as successful_requests,
    ROUND(COUNT(CASE WHEN status IN ('approved', 'completed') THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM requests 
WHERE created_at >= ?
GROUP BY CAST(strftime('%H', created_at) AS INTEGER)
ORDER BY hour_of_day;

-- GetContentAvailabilityVsRequests returns content popularity vs current availability
-- name: GetContentAvailabilityVsRequests :many
SELECT 
    r.tmdb_id,
    r.title,
    r.media_type,
    COUNT(r.id) as request_count,
    MAX(r.created_at) as last_requested,
    -- Check if content exists in library (basic availability check)
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM library_items li 
            WHERE li.tmdb_id = CAST(r.tmdb_id AS TEXT)
            AND (
                (r.media_type = 'movie' AND li.type = 'Movie') OR
                (r.media_type = 'tv' AND li.type IN ('Series', 'Season', 'Episode'))
            )
        ) THEN 1 
        ELSE 0 
    END as is_available,
    COUNT(CASE WHEN r.status IN ('approved', 'completed') THEN 1 END) as fulfilled_requests
FROM requests r
WHERE r.created_at >= ?
GROUP BY r.tmdb_id, r.title, r.media_type
ORDER BY COUNT(r.id) DESC, MAX(r.created_at) DESC
LIMIT ?;