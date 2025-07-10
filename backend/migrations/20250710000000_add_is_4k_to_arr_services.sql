-- Add is_4k column to arr_services table to identify 4K instances
ALTER TABLE arr_services ADD COLUMN is_4k BOOLEAN DEFAULT FALSE NOT NULL;