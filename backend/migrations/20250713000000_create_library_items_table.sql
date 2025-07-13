-- Migrate existing emby_media_items data if table exists
-- First, create the new improved library_items table
CREATE TABLE IF NOT EXISTS library_items (
    id TEXT PRIMARY KEY,                -- Emby/Jellyfin item ID
    name TEXT NOT NULL,                 -- Media item name/title
    original_title TEXT,                -- Original title (for international content)
    type TEXT NOT NULL,                 -- Media type (Movie, Series, Episode, Season, etc.)
    parent_id TEXT,                     -- Parent item ID (for episodes/seasons)
    series_id TEXT,                     -- Series ID for episodes
    season_number INTEGER,              -- Season number for episodes/seasons
    episode_number INTEGER,             -- Episode number for episodes
    year INTEGER,                       -- Release year
    premiere_date TEXT,                 -- Original premiere date (ISO format)
    end_date TEXT,                      -- End date for series
    community_rating REAL,              -- Community rating (like IMDb rating)
    critic_rating REAL,                 -- Critic rating
    official_rating TEXT,               -- Content rating (PG, R, etc.)
    overview TEXT,                      -- Plot summary/description
    tagline TEXT,                       -- Movie tagline
    genres TEXT,                        -- JSON array of genres
    studios TEXT,                       -- JSON array of studios
    people TEXT,                        -- JSON array of cast/crew
    tmdb_id TEXT,                       -- TMDB ID for matching with requests
    imdb_id TEXT,                       -- IMDB ID
    tvdb_id TEXT,                       -- TVDB ID
    musicbrainz_id TEXT,                -- MusicBrainz ID for music
    path TEXT,                          -- File path on server
    container TEXT,                     -- File container (mkv, mp4, etc.)
    size_bytes INTEGER,                 -- File size in bytes
    bitrate INTEGER,                    -- Video bitrate
    width INTEGER,                      -- Video width
    height INTEGER,                     -- Video height
    aspect_ratio TEXT,                  -- Video aspect ratio
    video_codec TEXT,                   -- Video codec
    audio_codec TEXT,                   -- Audio codec
    subtitle_tracks TEXT,               -- JSON array of subtitle tracks
    audio_tracks TEXT,                  -- JSON array of audio tracks
    runtime_ticks INTEGER,              -- Runtime in ticks
    runtime_minutes INTEGER,            -- Runtime in minutes (calculated)
    is_folder BOOLEAN DEFAULT FALSE,    -- Whether this is a folder/collection
    is_resumable BOOLEAN DEFAULT FALSE, -- Whether playback can be resumed
    play_count INTEGER DEFAULT 0,       -- Number of times played
    date_created TEXT,                  -- When item was added to library
    date_modified TEXT,                 -- When item was last modified
    last_played_date TEXT,              -- When item was last played
    user_data TEXT,                     -- JSON of user-specific data
    chapter_images_extracted BOOLEAN DEFAULT FALSE, -- Whether chapter images are available
    primary_image_tag TEXT,             -- Primary image tag/hash
    backdrop_image_tags TEXT,           -- JSON array of backdrop image tags
    logo_image_tag TEXT,                -- Logo image tag
    art_image_tag TEXT,                 -- Art image tag
    thumb_image_tag TEXT,               -- Thumbnail image tag
    is_hd BOOLEAN DEFAULT FALSE,        -- Whether content is HD
    is_4k BOOLEAN DEFAULT FALSE,        -- Whether content is 4K
    is_3d BOOLEAN DEFAULT FALSE,        -- Whether content is 3D
    locked BOOLEAN DEFAULT FALSE,       -- Whether metadata is locked
    provider_ids TEXT,                  -- JSON of all provider IDs
    external_urls TEXT,                 -- JSON of external URLs
    tags TEXT,                          -- JSON array of tags
    sort_name TEXT,                     -- Name used for sorting
    forced_sort_name TEXT,              -- Forced sort name
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Migrate existing data from old emby_media_items table if it exists
INSERT INTO library_items (
    id, name, type, year, tmdb_id, imdb_id, tvdb_id, path, runtime_ticks, updated_at, created_at
)
SELECT 
    id, 
    name, 
    type, 
    year, 
    tmdb_id, 
    imdb_id, 
    tvdb_id, 
    path, 
    runtime_ticks, 
    updated_at,
    COALESCE(updated_at, CURRENT_TIMESTAMP) as created_at
FROM emby_media_items 
WHERE EXISTS (SELECT name FROM sqlite_master WHERE type='table' AND name='emby_media_items');

-- Drop the old table after migration
DROP TABLE IF EXISTS emby_media_items;

-- Create view for backwards compatibility
CREATE VIEW IF NOT EXISTS emby_media_items AS 
SELECT 
    id,
    name,
    type,
    year,
    tmdb_id,
    imdb_id,
    tvdb_id,
    path,
    runtime_ticks,
    updated_at
FROM library_items;

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_library_items_tmdb_id ON library_items(tmdb_id) WHERE tmdb_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_imdb_id ON library_items(imdb_id) WHERE imdb_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_tvdb_id ON library_items(tvdb_id) WHERE tvdb_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_type ON library_items(type);
CREATE INDEX IF NOT EXISTS idx_library_items_name ON library_items(name);
CREATE INDEX IF NOT EXISTS idx_library_items_parent_id ON library_items(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_series_id ON library_items(series_id) WHERE series_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_year ON library_items(year) WHERE year IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_library_items_updated_at ON library_items(updated_at);
CREATE INDEX IF NOT EXISTS idx_library_items_date_created ON library_items(date_created);
CREATE INDEX IF NOT EXISTS idx_library_items_is_4k ON library_items(is_4k);
CREATE INDEX IF NOT EXISTS idx_library_items_is_hd ON library_items(is_hd);

-- Create a compound index for series episodes
CREATE INDEX IF NOT EXISTS idx_library_items_series_season_episode ON library_items(series_id, season_number, episode_number) WHERE type = 'Episode';