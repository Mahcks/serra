-- Add custom threshold columns to mounted_drives table
ALTER TABLE mounted_drives ADD COLUMN warning_threshold REAL DEFAULT 80.0;
ALTER TABLE mounted_drives ADD COLUMN critical_threshold REAL DEFAULT 95.0;
ALTER TABLE mounted_drives ADD COLUMN growth_rate_threshold REAL DEFAULT 50.0;