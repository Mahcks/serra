-- 1. Create a new table without user_id
CREATE TABLE arr_services_new (
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

-- 2. Copy data from old table to new table
INSERT INTO arr_services_new (
    id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability, created_at
)
SELECT
    id, type, name, base_url, api_key, quality_profile, root_folder_path, minimum_availability, created_at
FROM arr_services;

-- 3. Drop the old table
DROP TABLE arr_services;

-- 4. Rename the new table to the original name
ALTER TABLE arr_services_new RENAME TO arr_services;