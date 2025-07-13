-- Add auto-approval permissions to the permissions table
-- These allow users to automatically approve their own requests

-- Use INSERT OR IGNORE to prevent conflicts on re-runs
-- Auto-approval permissions
INSERT OR IGNORE INTO permissions (id, name, description) VALUES 
('request.auto_approve_movies', 'Auto-Approve Movies', 'Automatically approve movie requests'),
('request.auto_approve_series', 'Auto-Approve Series', 'Automatically approve TV series requests'),
('request.auto_approve_4k_movies', 'Auto-Approve 4K Movies', 'Automatically approve 4K movie requests'),
('request.auto_approve_4k_series', 'Auto-Approve 4K Series', 'Automatically approve 4K TV series requests');