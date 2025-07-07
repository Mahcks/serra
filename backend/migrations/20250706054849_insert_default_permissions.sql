-- Insert default permissions into the permissions table
-- These match the permissions defined in pkg/permissions/permissions.go

-- Use INSERT OR IGNORE to prevent conflicts on re-runs
-- Administrative permissions
INSERT OR IGNORE INTO permissions (id, name, description) VALUES 
('admin.users', 'Manage Users', 'Manage user accounts and permissions'),
('admin.services', 'Manage Services', 'Configure Radarr, Sonarr, and download clients'),
('admin.system', 'System Settings', 'Manage system settings, storage, and webhooks');

-- Request permissions
INSERT OR IGNORE INTO permissions (id, name, description) VALUES 
('request.movies', 'Request Movies', 'Submit movie requests'),
('request.series', 'Request Series', 'Submit TV series requests'),
('request.4k_movies', 'Request 4K Movies', 'Submit 4K movie requests'),
('request.4k_series', 'Request 4K Series', 'Submit 4K TV series requests');

-- Request management permissions
INSERT OR IGNORE INTO permissions (id, name, description) VALUES 
('requests.view', 'View All Requests', 'View all user requests'),
('requests.approve', 'Approve Requests', 'Approve or deny pending requests'),
('requests.manage', 'Manage Requests', 'Edit or delete any user requests');