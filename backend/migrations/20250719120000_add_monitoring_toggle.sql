-- Add monitoring_enabled column to mounted_drives table
ALTER TABLE mounted_drives ADD COLUMN monitoring_enabled BOOLEAN DEFAULT TRUE;