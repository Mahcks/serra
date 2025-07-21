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