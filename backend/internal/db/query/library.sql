-- name: CheckMediaInLibrary :one
SELECT COUNT(*) > 0 as in_library 
FROM library_items 
WHERE tmdb_id = ?;

-- name: CheckMultipleMediaInLibrary :many
SELECT tmdb_id, COUNT(*) > 0 as in_library
FROM library_items 
WHERE tmdb_id IN (/*SLICE:tmdb_ids*/?)
GROUP BY tmdb_id;

-- name: GetLibraryItemByTMDBID :one
SELECT id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date, 
       community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
       tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height, 
       aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
       is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
       chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag, 
       thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name, 
       forced_sort_name, created_at, updated_at
FROM library_items
WHERE tmdb_id = ?
LIMIT 1;

-- name: SearchLibraryByTitle :many
SELECT id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date, 
       community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
       tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height, 
       aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
       is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
       chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag, 
       thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name, 
       forced_sort_name, created_at, updated_at
FROM library_items
WHERE name LIKE '%' || ? || '%'
ORDER BY name
LIMIT ?;

-- name: CreateLibraryItemFull :one
INSERT INTO library_items (
    id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date, 
    community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
    tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height, 
    aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
    is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
    chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag, 
    thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name, 
    forced_sort_name, created_at, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date, 
          community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
          tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height, 
          aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
          is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
          chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag, 
          thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name, 
          forced_sort_name, created_at, updated_at;

-- name: CreateLibraryItem :one
INSERT INTO library_items (
    id, name, type, year, tmdb_id, imdb_id, tvdb_id, path, runtime_ticks, updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, original_title, type, parent_id, series_id, season_number, episode_number, year, premiere_date, end_date, 
          community_rating, critic_rating, official_rating, overview, tagline, genres, studios, people,
          tmdb_id, imdb_id, tvdb_id, musicbrainz_id, path, container, size_bytes, bitrate, width, height, 
          aspect_ratio, video_codec, audio_codec, subtitle_tracks, audio_tracks, runtime_ticks, runtime_minutes,
          is_folder, is_resumable, play_count, date_created, date_modified, last_played_date, user_data,
          chapter_images_extracted, primary_image_tag, backdrop_image_tags, logo_image_tag, art_image_tag, 
          thumb_image_tag, is_hd, is_4k, is_3d, locked, provider_ids, external_urls, tags, sort_name, 
          forced_sort_name, created_at, updated_at;

-- Compatibility query for legacy CreateEmbyMediaItem usage
-- name: CreateEmbyMediaItem :one
INSERT INTO library_items (id, name, type, year, tmdb_id, imdb_id, tvdb_id, path, runtime_ticks, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id, name, type, year, tmdb_id, imdb_id, tvdb_id, path, runtime_ticks, updated_at;