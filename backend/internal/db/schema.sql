CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    access_token TEXT,
    avatar_url TEXT,
    email TEXT,
    user_type TEXT DEFAULT 'media_server' NOT NULL,
    password_hash TEXT,
    invited_by TEXT,
    invitation_accepted_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

-- Invitations table for invitation system
CREATE TABLE invitations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    invited_by TEXT NOT NULL,
    permissions TEXT, -- JSON array of permission IDs
    create_media_user BOOLEAN DEFAULT true,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'cancelled')),
    expires_at DATETIME NOT NULL,
    accepted_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (invited_by) REFERENCES users(id)
);

-- Indexes for invitations table
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_status ON invitations(status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

CREATE TABLE downloads (
    id TEXT PRIMARY KEY,
    -- can be downloadId from qBittorrent or constructed from source + ID
    title TEXT NOT NULL,
    torrent_title TEXT NOT NULL,
    source TEXT NOT NULL,
    -- "radarr", "sonarr", or "qbittorrent"
    tmdb_id INTEGER,
    -- for Radarr
    tvdb_id INTEGER,
    -- for Sonarr
    hash TEXT,
    -- for qBittorrent
    progress REAL,
    time_left TEXT,
    status TEXT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    download_speed INTEGER, -- bytes per second
    upload_speed INTEGER,   -- bytes per second
    download_size INTEGER   -- total download size in bytes
);

CREATE TABLE IF NOT EXISTS service_status (
    name TEXT PRIMARY KEY,
    online BOOLEAN,
    last_checked TIMESTAMP
);

CREATE TABLE library_items (
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

CREATE TABLE IF NOT EXISTS user_settings (
    user_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, key)
);

CREATE TABLE IF NOT EXISTS requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    -- Who made the request (foreign key to users)
    media_type TEXT NOT NULL,
    -- 'movie' or 'series'
    tmdb_id INTEGER,
    -- TMDB ID (preferred)
    title TEXT,
    -- Requested title
    status TEXT NOT NULL,
    -- 'pending', 'approved', 'denied', 'fulfilled', etc.
    notes TEXT,
    -- Optional message/request details
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fulfilled_at DATETIME,
    -- When it was actually fulfilled (nullable)
    approver_id TEXT,
    -- Optional: Who approved/denied
    on_behalf_of TEXT,
    -- Optional: User on whose behalf this request was made
    poster_url TEXT,
    -- Optional: URL to a poster image for the request
    seasons TEXT DEFAULT NULL,
    -- For TV shows - JSON array of season numbers being requested
    season_statuses TEXT DEFAULT NULL,
    -- JSON object tracking individual season statuses
    UNIQUE (media_type, tmdb_id, user_id, seasons) -- Allow different season combinations per user
);

-- Table for tracking availability of TV show seasons in media server
CREATE TABLE season_availability (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tmdb_id INTEGER NOT NULL,
    season_number INTEGER NOT NULL,
    episode_count INTEGER NOT NULL,
    available_episodes INTEGER DEFAULT 0,
    is_complete BOOLEAN DEFAULT false,
    last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tmdb_id, season_number)
);

CREATE TABLE download_clients (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('qbittorrent', 'sabnzbd')),
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT,
    password TEXT,
    api_key TEXT,
    -- only used for sabnzbd
    use_ssl BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE arr_services (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('sonarr', 'radarr')),
    name TEXT NOT NULL,
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL,
    quality_profile TEXT NOT NULL,
    root_folder_path TEXT NOT NULL,
    minimum_availability TEXT NOT NULL,
    is_4k BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE mounted_drives (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    mount_path TEXT NOT NULL UNIQUE,
    -- The filesystem mount point (e.g., /mnt/media, /media/storage)
    filesystem TEXT,
    -- Filesystem type (e.g., ext4, ntfs, zfs)
    total_size BIGINT,
    -- Total size in bytes
    used_size BIGINT,
    -- Used size in bytes
    available_size BIGINT,
    -- Available size in bytes
    usage_percentage REAL,
    -- Usage percentage (0-100)
    is_online BOOLEAN DEFAULT TRUE,
    -- Whether the drive is currently accessible
    monitoring_enabled BOOLEAN DEFAULT TRUE,
    -- Whether monitoring is enabled for this drive
    warning_threshold REAL DEFAULT 80.0,
    -- Warning threshold percentage (default 80%)
    critical_threshold REAL DEFAULT 95.0,
    -- Critical threshold percentage (default 95%)
    growth_rate_threshold REAL DEFAULT 50.0,
    -- Growth rate threshold in GB per day (default 50GB)
    last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permissions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE user_permissions (
    user_id TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Default permissions table for storing which permissions are enabled by default for new users
CREATE TABLE default_permissions (
    permission_id TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- TMDB Cache Tables
CREATE TABLE tmdb_cache (
    cache_key TEXT PRIMARY KEY,
    -- Format: "tmdb:{endpoint}:{params_hash}"
    data TEXT NOT NULL,
    -- JSON data from TMDB API
    endpoint TEXT NOT NULL,
    -- e.g., "discover/movie", "search/company", "movie/{id}"
    expires_at TIMESTAMP NOT NULL,
    -- When this cache entry expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient cleanup of expired entries
CREATE INDEX idx_tmdb_cache_expires_at ON tmdb_cache(expires_at);
CREATE INDEX idx_tmdb_cache_endpoint ON tmdb_cache(endpoint);

-- Separate table for static TMDB data (genres, companies, etc.)
CREATE TABLE tmdb_static_data (
    data_type TEXT PRIMARY KEY,
    -- e.g., "genres", "companies", "watch_providers"
    data TEXT NOT NULL,
    -- JSON data
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table for tracking TMDB API usage and rate limiting
CREATE TABLE tmdb_api_usage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    endpoint TEXT NOT NULL,
    request_count INTEGER DEFAULT 1,
    date DATE NOT NULL,
    -- Date of the requests (for daily tracking)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for efficient API usage tracking
CREATE UNIQUE INDEX idx_tmdb_api_usage_endpoint_date ON tmdb_api_usage(endpoint, date);

-- Analytics tables for request system operations and drive monitoring

-- Request analytics for tracking popularity and processing metrics
CREATE TABLE request_analytics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tmdb_id INTEGER NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    title TEXT NOT NULL,
    request_count INTEGER DEFAULT 1,
    last_requested TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    first_requested TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avg_processing_time_seconds INTEGER DEFAULT 0,
    success_rate REAL DEFAULT 0.0, -- 0.0 to 1.0
    popularity_score REAL DEFAULT 0.0, -- Calculated popularity metric
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tmdb_id, media_type)
);

-- Individual request metrics for detailed tracking
CREATE TABLE request_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER NOT NULL,
    status_change TEXT NOT NULL, -- 'created', 'approved', 'processing', 'fulfilled', 'failed'
    previous_status TEXT,
    new_status TEXT NOT NULL,
    processing_time_seconds INTEGER, -- Time taken for this status change
    error_code INTEGER, -- API error code if applicable
    error_message TEXT, -- Error details if applicable
    user_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE
);

-- Drive usage history for monitoring and alerting
CREATE TABLE drive_usage_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    drive_id TEXT NOT NULL,
    total_size BIGINT NOT NULL,
    used_size BIGINT NOT NULL,
    available_size BIGINT NOT NULL,
    usage_percentage REAL NOT NULL,
    growth_rate_gb_per_day REAL DEFAULT 0.0, -- How fast the drive is filling up
    projected_full_date DATE, -- When the drive is estimated to be full
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (drive_id) REFERENCES mounted_drives(id) ON DELETE CASCADE
);

-- Drive monitoring alerts and thresholds
CREATE TABLE drive_alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    drive_id TEXT NOT NULL,
    alert_type TEXT NOT NULL CHECK (alert_type IN ('usage_threshold', 'growth_rate', 'projected_full')),
    threshold_value REAL NOT NULL, -- e.g., 80.0 for 80% usage threshold
    current_value REAL NOT NULL,
    alert_message TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    acknowledgement_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (drive_id) REFERENCES mounted_drives(id) ON DELETE CASCADE
);

-- System analytics for overall health and performance
CREATE TABLE system_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_type TEXT NOT NULL, -- 'request_volume', 'processing_performance', 'service_health'
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    metadata TEXT, -- JSON for additional context
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Popular content forecasting based on external trends
CREATE TABLE popularity_trends (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tmdb_id INTEGER NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    title TEXT NOT NULL,
    trend_source TEXT NOT NULL, -- 'tmdb_trending', 'request_pattern', 'seasonal'
    popularity_score REAL NOT NULL,
    trend_direction TEXT CHECK (trend_direction IN ('rising', 'stable', 'declining')),
    forecast_confidence REAL DEFAULT 0.0, -- 0.0 to 1.0
    metadata TEXT, -- JSON for additional trend data
    valid_until TIMESTAMP, -- When this trend data expires
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tmdb_id, media_type, trend_source)
);

-- Indexes for efficient querying
CREATE INDEX idx_request_analytics_popularity ON request_analytics(popularity_score DESC);
CREATE INDEX idx_request_analytics_media_type ON request_analytics(media_type);
CREATE INDEX idx_request_metrics_request_id ON request_metrics(request_id);
CREATE INDEX idx_request_metrics_timestamp ON request_metrics(timestamp);
CREATE INDEX idx_drive_usage_history_drive_id ON drive_usage_history(drive_id);
CREATE INDEX idx_drive_usage_history_recorded_at ON drive_usage_history(recorded_at);
CREATE INDEX idx_drive_alerts_active ON drive_alerts(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_system_metrics_type_name ON system_metrics(metric_type, metric_name);
CREATE INDEX idx_popularity_trends_score ON popularity_trends(popularity_score DESC);
CREATE INDEX idx_popularity_trends_media_type ON popularity_trends(media_type);