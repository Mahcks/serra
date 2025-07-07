-- Insert owner permission into the permissions table
INSERT OR IGNORE INTO permissions (id, name, description) VALUES 
('owner', 'Owner', 'Full system access - cannot be revoked');