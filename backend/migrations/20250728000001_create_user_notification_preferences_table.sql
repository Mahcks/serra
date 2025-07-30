-- Create user notification preferences table
CREATE TABLE user_notification_preferences (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    
    -- Individual notification type preferences
    requests_approved BOOLEAN DEFAULT TRUE,
    requests_denied BOOLEAN DEFAULT TRUE,
    download_completed BOOLEAN DEFAULT TRUE,
    media_available BOOLEAN DEFAULT TRUE,
    system_alerts BOOLEAN DEFAULT TRUE,
    
    -- Priority level preferences (minimum priority to receive)
    min_priority TEXT DEFAULT 'low' CHECK (min_priority IN ('low', 'normal', 'high', 'urgent')),
    
    -- Delivery preferences
    web_notifications BOOLEAN DEFAULT TRUE,
    email_notifications BOOLEAN DEFAULT FALSE,
    push_notifications BOOLEAN DEFAULT FALSE,
    
    -- Time-based preferences
    quiet_hours_enabled BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME, -- Start of quiet hours (24-hour format)
    quiet_hours_end TIME,   -- End of quiet hours (24-hour format)
    
    -- Auto-cleanup preferences
    auto_mark_read_after_days INTEGER DEFAULT NULL, -- Auto-mark as read after X days (NULL = never)
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for user notification preferences
CREATE INDEX idx_user_notification_preferences_user_id ON user_notification_preferences(user_id);

-- Trigger to update updated_at timestamp
CREATE TRIGGER update_user_notification_preferences_updated_at
    AFTER UPDATE ON user_notification_preferences
    FOR EACH ROW
BEGIN
    UPDATE user_notification_preferences 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;

-- Create default notification preferences for existing users
INSERT INTO user_notification_preferences (id, user_id)
SELECT 
    lower(hex(randomblob(16))),
    id
FROM users
WHERE id NOT IN (SELECT user_id FROM user_notification_preferences);