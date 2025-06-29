-- 1. Create a new table without user_id
CREATE TABLE download_clients_new (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('qbittorrent', 'sabnzbd')),
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL,
    username TEXT,
    password TEXT,
    api_key TEXT,
    use_ssl BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Copy data from old table to new table
INSERT INTO download_clients_new (
    id, type, name, host, port, username, password, api_key, use_ssl, created_at
)
SELECT
    id, type, name, host, port, username, password, api_key, use_ssl, created_at
FROM download_clients;

-- 3. Drop the old table
DROP TABLE download_clients;

-- 4. Rename the new table to the original name
ALTER TABLE download_clients_new RENAME TO download_clients; 