CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    access_token TEXT,
    email TEXT
);

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
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS service_status (
    name TEXT PRIMARY KEY,
    online BOOLEAN,
    last_checked TIMESTAMP
);

CREATE TABLE emby_media_items (
    id TEXT PRIMARY KEY,
    -- Emby's internal item ID (unique per Emby server)
    name TEXT NOT NULL,
    -- Title of the movie/show
    type TEXT NOT NULL,
    -- e.g. 'Movie', 'Series'
    year INTEGER,
    -- Release year (if available)
    tmdb_id TEXT,
    -- TMDb ID (for existence checks)
    imdb_id TEXT,
    -- IMDb ID (for existence checks)
    tvdb_id TEXT,
    -- TVDb ID (for TV shows)
    path TEXT,
    -- Local filesystem path (optional, but useful for admin)
    runtime_ticks BIGINT,
    -- Emby’s internal runtime value (optional)
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP -- When we last synced this row
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
    UNIQUE (media_type, tmdb_id, user_id) -- Prevent duplicate requests from same user for same item
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
    last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);