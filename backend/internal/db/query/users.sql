-- name: CreateUser :one
INSERT INTO users (id, username, access_token, email, avatar_url, user_type, password_hash)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  username = excluded.username,
  access_token = excluded.access_token,
  avatar_url = excluded.avatar_url,
  user_type = excluded.user_type,
  updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UserExists :one
SELECT EXISTS(
  SELECT 1 FROM users WHERE id = ?
);

-- name: GetAllUsers :many
SELECT id, username, email, avatar_url, user_type, created_at FROM users;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? AND user_type = 'local';

-- name: CreateLocalUser :one
INSERT INTO users (id, username, email, password_hash, user_type, avatar_url)
VALUES (?, ?, ?, ?, 'local', ?)
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? AND user_type = 'local';
