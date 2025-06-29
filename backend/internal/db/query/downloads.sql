-- name: UpsertDownloadQueue :exec
INSERT INTO downloads (
  id, title, torrent_title, source, tmdb_id, tvdb_id, hash, progress, time_left, status, last_updated
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(id) DO UPDATE SET
  title = excluded.title,
  torrent_title = excluded.torrent_title,
  source = excluded.source,
  tmdb_id = excluded.tmdb_id,
  tvdb_id = excluded.tvdb_id,
  hash = excluded.hash,
  progress = excluded.progress,
  time_left = excluded.time_left,
  status = excluded.status,
  last_updated = CURRENT_TIMESTAMP;

-- name: ListDownloads :many
SELECT
  id,
  title,
  torrent_title,
  source,
  tmdb_id,
  tvdb_id,
  hash,
  progress,
  time_left,
  status,
  last_updated
FROM downloads
WHERE status IS NULL OR status != 'completed'
ORDER BY last_updated DESC;

-- name: ListDownloadsBySource :many
SELECT * FROM downloads WHERE source = ?;

-- name: DeleteDownload :exec
DELETE FROM downloads WHERE id = ?;

-- name: GetOldMissingDownloads :many
SELECT
  id,
  title,
  torrent_title,
  source,
  tmdb_id,
  tvdb_id,
  hash,
  progress,
  time_left,
  status,
  last_updated
FROM downloads
WHERE status = 'missing_from_client'
  AND last_updated < datetime('now', '-24 hours')
ORDER BY last_updated ASC;
