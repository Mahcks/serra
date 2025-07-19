-- Add season-level request tracking

-- Add columns to requests table for season tracking  
ALTER TABLE requests ADD COLUMN seasons TEXT DEFAULT NULL;
ALTER TABLE requests ADD COLUMN season_statuses TEXT DEFAULT NULL;

-- Create season_availability table for tracking what's available in media server
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

-- Create indexes for better performance
CREATE INDEX idx_season_availability_tmdb ON season_availability(tmdb_id);
CREATE INDEX idx_season_availability_complete ON season_availability(is_complete);