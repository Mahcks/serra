ALTER TABLE downloads ADD COLUMN download_speed INTEGER; -- bytes per second
ALTER TABLE downloads ADD COLUMN upload_speed INTEGER;   -- bytes per second  
ALTER TABLE downloads ADD COLUMN download_size INTEGER;  -- total download size in bytes