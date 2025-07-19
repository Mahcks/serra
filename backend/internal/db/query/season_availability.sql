-- name: GetSeasonAvailabilityByTMDBID :many
SELECT id, tmdb_id, season_number, episode_count, available_episodes, is_complete, last_updated
FROM season_availability
WHERE tmdb_id = ?
ORDER BY season_number;

-- name: GetSeasonAvailabilityByTMDBIDAndSeason :one
SELECT id, tmdb_id, season_number, episode_count, available_episodes, is_complete, last_updated
FROM season_availability
WHERE tmdb_id = ? AND season_number = ?;

-- name: UpsertSeasonAvailability :exec
INSERT INTO season_availability (tmdb_id, season_number, episode_count, available_episodes, is_complete, last_updated)
VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(tmdb_id, season_number) DO UPDATE SET
    episode_count = excluded.episode_count,
    available_episodes = excluded.available_episodes,
    is_complete = excluded.is_complete,
    last_updated = CURRENT_TIMESTAMP;

-- name: UpdateSeasonAvailableEpisodes :exec
INSERT INTO season_availability (tmdb_id, season_number, episode_count, available_episodes, last_updated)
VALUES (?, ?, 0, ?, CURRENT_TIMESTAMP)
ON CONFLICT(tmdb_id, season_number) DO UPDATE SET
    available_episodes = excluded.available_episodes,
    is_complete = (excluded.available_episodes >= season_availability.episode_count AND season_availability.episode_count > 0),
    last_updated = CURRENT_TIMESTAMP;