-- name: GetRequestAnalytics :many
SELECT * FROM request_analytics 
ORDER BY popularity_score DESC, request_count DESC
LIMIT ?;

-- name: GetRequestAnalyticsByMediaType :many
SELECT * FROM request_analytics 
WHERE media_type = ?
ORDER BY popularity_score DESC, request_count DESC
LIMIT ?;

-- name: UpsertRequestAnalytics :one
INSERT INTO request_analytics (tmdb_id, media_type, title, request_count, last_requested, first_requested, avg_processing_time_seconds, success_rate, popularity_score)
VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
ON CONFLICT(tmdb_id, media_type) DO UPDATE SET
    request_count = request_count + 1,
    last_requested = CURRENT_TIMESTAMP,
    avg_processing_time_seconds = (avg_processing_time_seconds * (request_count - 1) + ?) / request_count,
    success_rate = ?,
    popularity_score = ?,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: RecordRequestMetric :one
INSERT INTO request_metrics (request_id, status_change, previous_status, new_status, processing_time_seconds, error_code, error_message, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetRequestMetricsByRequestID :many
SELECT * FROM request_metrics 
WHERE request_id = ?
ORDER BY timestamp ASC;

-- name: GetRecentRequestMetrics :many
SELECT * FROM request_metrics 
WHERE timestamp >= datetime('now', '-' || ? || ' days')
ORDER BY timestamp DESC
LIMIT ?;

-- name: RecordDriveUsage :one
INSERT INTO drive_usage_history (drive_id, total_size, used_size, available_size, usage_percentage, growth_rate_gb_per_day, projected_full_date)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetDriveUsageHistory :many
SELECT * FROM drive_usage_history 
WHERE drive_id = ?1
ORDER BY recorded_at DESC
LIMIT ?2;

-- name: GetLatestDriveUsage :many
SELECT 
    duh.*,
    md.name as drive_name,
    md.mount_path as drive_mount_path
FROM drive_usage_history duh
INNER JOIN (
    SELECT drive_id, MAX(recorded_at) as latest_recorded
    FROM drive_usage_history
    GROUP BY drive_id
) latest ON duh.drive_id = latest.drive_id AND duh.recorded_at = latest.latest_recorded
LEFT JOIN mounted_drives md ON duh.drive_id = md.id
ORDER BY duh.usage_percentage DESC;

-- name: CreateDriveAlert :one
INSERT INTO drive_alerts (drive_id, alert_type, threshold_value, current_value, alert_message)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetActiveDriveAlerts :many
SELECT da.*, md.name as drive_name, md.mount_path 
FROM drive_alerts da
JOIN mounted_drives md ON da.drive_id = md.id
WHERE da.is_active = TRUE
ORDER BY da.last_triggered DESC;

-- name: DeactivateDriveAlert :exec
UPDATE drive_alerts 
SET is_active = FALSE, acknowledgement_count = acknowledgement_count + 1
WHERE id = ?;

-- name: ClearDriveAlerts :exec
UPDATE drive_alerts 
SET is_active = FALSE
WHERE drive_id = ? AND is_active = TRUE;

-- name: RecordSystemMetric :one
INSERT INTO system_metrics (metric_type, metric_name, metric_value, metadata)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetSystemMetrics :many
SELECT * FROM system_metrics 
WHERE metric_type = ? AND recorded_at >= datetime('now', '-' || ? || ' days')
ORDER BY recorded_at DESC
LIMIT ?;

-- name: UpsertPopularityTrend :one
INSERT INTO popularity_trends (tmdb_id, media_type, title, trend_source, popularity_score, trend_direction, forecast_confidence, metadata, valid_until)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(tmdb_id, media_type, trend_source) DO UPDATE SET
    title = ?,
    popularity_score = ?,
    trend_direction = ?,
    forecast_confidence = ?,
    metadata = ?,
    valid_until = ?,
    created_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetTrendingContent :many
SELECT * FROM popularity_trends 
WHERE valid_until > CURRENT_TIMESTAMP
ORDER BY popularity_score DESC, forecast_confidence DESC
LIMIT ?;

-- name: GetTrendingContentByType :many
SELECT * FROM popularity_trends 
WHERE media_type = ? AND valid_until > CURRENT_TIMESTAMP
ORDER BY popularity_score DESC, forecast_confidence DESC
LIMIT ?;

-- name: GetRequestSuccessRate :one
SELECT 
    COUNT(*) as total_requests,
    SUM(CASE WHEN rm.new_status = 'fulfilled' THEN 1 ELSE 0 END) as fulfilled_requests,
    CAST(SUM(CASE WHEN rm.new_status = 'fulfilled' THEN 1 ELSE 0 END) AS REAL) / COUNT(*) as success_rate
FROM request_metrics rm
WHERE rm.timestamp >= datetime('now', '-' || ? || ' days')
AND rm.status_change = 'created';

-- name: GetAverageProcessingTime :one
SELECT 
    AVG(
        CASE 
            WHEN rm_end.timestamp IS NOT NULL THEN 
                (julianday(rm_end.timestamp) - julianday(rm_start.timestamp)) * 24 * 60 * 60
            ELSE NULL 
        END
    ) as avg_processing_seconds
FROM request_metrics rm_start
LEFT JOIN request_metrics rm_end ON rm_start.request_id = rm_end.request_id 
    AND rm_end.new_status IN ('fulfilled', 'failed')
WHERE rm_start.status_change = 'created'
AND rm_start.timestamp >= datetime('now', '-' || ? || ' days');

-- name: GetMostRequestedContent :many
SELECT 
    tmdb_id,
    media_type,
    title,
    request_count,
    last_requested,
    popularity_score
FROM request_analytics
WHERE last_requested >= datetime('now', '-' || ? || ' days')
ORDER BY request_count DESC
LIMIT ?;

-- name: GetDriveGrowthPrediction :many
SELECT 
    drive_id,
    AVG(growth_rate_gb_per_day) as avg_growth_rate_gb_per_day,
    AVG(usage_percentage) as current_avg_usage,
    MIN(projected_full_date) as earliest_full_date
FROM drive_usage_history
WHERE recorded_at >= datetime('now', '-' || ?1 || ' days')
GROUP BY drive_id
HAVING COUNT(*) >= 2
ORDER BY current_avg_usage DESC;