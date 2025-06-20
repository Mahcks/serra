-- name: CreateUser :one
INSERT INTO users (id, username, access_token, email)
VALUES (?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  username = excluded.username,
  access_token = excluded.access_token
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: UserExists :one
SELECT EXISTS(
  SELECT 1 FROM users WHERE id = ?
);

-- name: GetAllUsers :many
SELECT id, username, email FROM users;
