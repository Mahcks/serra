-- Create mounted_drives table
CREATE TABLE mounted_drives (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    mount_path TEXT NOT NULL UNIQUE,
    -- The filesystem mount point (e.g., /mnt/media, /media/storage)
    filesystem TEXT,
    -- Filesystem type (e.g., ext4, ntfs, zfs)
    total_size BIGINT,
    -- Total size in bytes
    used_size BIGINT,
    -- Used size in bytes
    available_size BIGINT,
    -- Available size in bytes
    usage_percentage REAL,
    -- Usage percentage (0-100)
    is_online BOOLEAN DEFAULT TRUE,
    -- Whether the drive is currently accessible
    last_checked TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);