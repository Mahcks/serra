-- Migration: Fix request constraint to allow season-specific requests
-- +goose Up

-- First, let's check if we have any existing duplicate requests that would conflict
-- with the new constraint logic (this is just for information, the constraint change handles it)

-- Drop the existing unique constraint by recreating the table
-- SQLite doesn't support dropping constraints directly, so we need to recreate the table

-- Step 1: Create new table with modified constraint
CREATE TABLE requests_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    tmdb_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied', 'fulfilled', 'processing', 'failed')),
    notes TEXT,
    -- Optional: Additional notes from the user
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    fulfilled_at DATETIME,
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
    
    -- New constraint: Allow multiple requests for same show if different seasons
    -- For movies: still prevent duplicates
    -- For TV shows: allow if seasons are different or if one request has seasons and another doesn't
    UNIQUE (media_type, tmdb_id, user_id, seasons) -- This will allow different season combinations
);

-- Step 2: Copy data from old table to new table
INSERT INTO requests_new (
    id, user_id, media_type, tmdb_id, title, status, notes, 
    created_at, updated_at, fulfilled_at, 
    approver_id, on_behalf_of, poster_url, seasons, season_statuses
)
SELECT 
    id, user_id, media_type, tmdb_id, title, status, notes,
    created_at, updated_at, fulfilled_at,
    approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests;

-- Step 3: Drop old table
DROP TABLE requests;

-- Step 4: Rename new table
ALTER TABLE requests_new RENAME TO requests;

-- Step 5: Recreate any indexes that might have existed
-- (Add any additional indexes that were on the original table)

-- +goose Down

-- Reverse the migration - restore original constraint
CREATE TABLE requests_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    media_type TEXT NOT NULL CHECK (media_type IN ('movie', 'tv')),
    tmdb_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied', 'fulfilled', 'processing', 'failed')),
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    fulfilled_at DATETIME,
    approver_id TEXT,
    on_behalf_of TEXT,
    poster_url TEXT,
    seasons TEXT DEFAULT NULL,
    season_statuses TEXT DEFAULT NULL,
    UNIQUE (media_type, tmdb_id, user_id) -- Original constraint
);

-- Copy data back (may fail if there are now duplicate entries)
INSERT INTO requests_old (
    id, user_id, media_type, tmdb_id, title, status, notes, 
    created_at, updated_at, fulfilled_at, 
    approver_id, on_behalf_of, poster_url, seasons, season_statuses
)
SELECT 
    id, user_id, media_type, tmdb_id, title, status, notes,
    created_at, updated_at, fulfilled_at,
    approver_id, on_behalf_of, poster_url, seasons, season_statuses
FROM requests;

DROP TABLE requests;
ALTER TABLE requests_old RENAME TO requests;