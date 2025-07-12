-- Add TMDB Cache Tables
-- Create Date: 2025-01-12
-- Description: Add comprehensive TMDB caching system with cache entries, static data, and API usage tracking

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