-- Create table for storing which permissions are enabled by default for new users
-- +migrate Up
CREATE TABLE IF NOT EXISTS default_permissions (
    permission_id TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert all current permissions with default false
-- Owner permission (usually not recommended as default)
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('owner', FALSE);

-- Admin permissions
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('admin.users', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('admin.services', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('admin.system', FALSE);

-- Request permissions
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.movies', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.series', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.4k_movies', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.4k_series', FALSE);

-- Auto-approval permissions
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.auto_approve_movies', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.auto_approve_series', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.auto_approve_4k_movies', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('request.auto_approve_4k_series', FALSE);

-- Request management permissions
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('requests.view', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('requests.approve', FALSE);
INSERT OR IGNORE INTO default_permissions (permission_id, enabled) VALUES ('requests.manage', FALSE);