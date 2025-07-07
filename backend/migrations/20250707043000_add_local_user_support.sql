-- Add support for local users
ALTER TABLE users ADD COLUMN user_type TEXT DEFAULT 'media_server' NOT NULL;
ALTER TABLE users ADD COLUMN password_hash TEXT;
ALTER TABLE users ADD COLUMN created_at DATETIME;
ALTER TABLE users ADD COLUMN updated_at DATETIME;

-- Update existing users with current timestamp and set to media_server type
UPDATE users SET 
  user_type = 'media_server',
  created_at = CURRENT_TIMESTAMP,
  updated_at = CURRENT_TIMESTAMP 
WHERE user_type IS NULL OR created_at IS NULL;